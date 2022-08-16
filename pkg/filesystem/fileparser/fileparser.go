package fileparser

import (
	"bufio"
	"io/fs"
	"monster/pkg/common"
	"monster/pkg/utils/parsing"
	"os"
	"strings"
)

// APPEND: combine files (需要合并的文件)
// section: 文件里: [section]

type FileParser struct {
	isModFile    bool
	filenames    []string
	currentIndex uint32
	lineNumber   uint32
	infile       *os.File
	includeFp    *FileParser
	section      string
	newSection   bool
	key          string
	val          string
	scanner      *bufio.Scanner
}

func New() *FileParser {
	f := Construct()
	return &f
}

func Construct() FileParser {
	return FileParser{}
}

// 打开指定的文件或已加载mod内指定的全部文件
func (this *FileParser) Open(filename string, isModFile bool, mods common.ModManager) error {
	this.isModFile = isModFile
	var err error

	this.filenames = nil
	if this.isModFile {
		// 列出指定的mod内的某个文件 (肯定存在！)	[default, ...]
		this.filenames, err = mods.List(filename)
		if err != nil {
			return err
		}
	} else {
		// (不一定存在！)
		this.filenames = append(this.filenames, filename)
	}
	this.currentIndex = 0
	this.lineNumber = 0

	//fmt.Println(filename, this.filenames)
	if len(this.filenames) == 0 {
		return fs.ErrNotExist
	}

	// [default, ...]
	// 逆序从其他mod到default 找到文件位置
	// 找到第一个非append的文件
	for i := len(this.filenames); i > 0; i-- {
		if f, err := os.Open(this.filenames[i-1]); err == nil {
			this.infile = f
			scanner := bufio.NewScanner(this.infile)
			if scanner.Scan() && strings.TrimSpace(scanner.Text()) != "APPEND" {
				testLine := ""

				// get the first non-comment, non blank line
				// 获取第一个非空，非注释的行
				for scanner.Scan() {
					testLine = strings.TrimSpace(scanner.Text())
					if len(testLine) == 0 {
						continue
					} else if strings.HasPrefix(testLine, "#") {
						continue
					} else {
						break
					}
				}

				if testLine != "APPEND" {
					this.currentIndex = (uint32)(i - 1)
					if _, err = this.infile.Seek(0, 0); err != nil {
						this.infile.Close()
						this.infile = nil
						return err
					}

					break
				}
			}

			// APPEND file or normal file
			// don't close the final file if it's the only one with an "APPEND" line
			if i > 1 {
				this.infile.Close()
				this.infile = nil
			}

		} else {
			return err
		}

	}

	this.scanner = bufio.NewScanner(this.infile)
	return nil
}

func (this *FileParser) Close() {
	if this.includeFp != nil {
		this.includeFp.Close()
		this.includeFp = nil
	}

	if this.infile != nil {
		this.infile.Close()
		this.infile = nil
	}

	this.filenames = nil
	this.scanner = nil
}

// [default, ...]
// 顺序从文件位置到mod列表的末尾
func (this *FileParser) Next(mods common.ModManager) bool {
	if this.infile == nil {
		return false
	}

	this.newSection = false

	for (int)(this.currentIndex) < len(this.filenames) {

		scanOk := true
		for scanOk {
			if this.includeFp != nil {
				if this.includeFp.Next(mods) {
					this.newSection = this.includeFp.newSection
					this.section = this.includeFp.section
					this.key = this.includeFp.key
					this.val = this.includeFp.val
					return true
				} else {
					this.includeFp.Close()
					this.includeFp = nil
					continue
				}
			}

			//for
			scanOk = this.scanner.Scan()
			if !scanOk {
				continue
			}
			line := strings.TrimSpace(this.scanner.Text())
			this.lineNumber++

			if len(line) == 0 {
				continue
			}

			if strings.HasPrefix(line, "#") {
				continue
			}

			if strings.HasPrefix(line, "[") {
				this.newSection = true
				this.section = parsing.GetSectionTitle(line)
				continue
			}

			if line == "APPEND" {
				continue
			}

			firstSpace := strings.Index(line, " ")
			if firstSpace > 0 {
				directive := strings.TrimSpace(line[0:firstSpace])

				if directive == "INCLUDE" {
					tmp := line[firstSpace+1:]
					this.includeFp = New()
					if err := this.includeFp.Open(tmp, this.isModFile, mods); err != nil {
						this.includeFp.Close()
						this.includeFp = nil
						panic(err)
					}

					this.includeFp.section = this.section
					continue
				}
			}

			this.key, this.val = parsing.GetKeyPair(this.scanner.Bytes())
			return true
		}

		if err := this.scanner.Err(); err != nil {
			panic(err)
		}

		this.infile.Close()
		this.infile = nil
		this.scanner = nil
		this.currentIndex++

		if (int)(this.currentIndex) == len(this.filenames) {
			return false
		}

		this.lineNumber = 0
		currentFilename := this.filenames[this.currentIndex]

		if f, err := os.Open(currentFilename); err != nil {
			return false
		} else {
			this.infile = f
			this.scanner = bufio.NewScanner(this.infile)
		}

		// a new file starts a new section
		this.newSection = true
	}

	return false
}

func (this *FileParser) Key() string {
	return this.key
}

func (this *FileParser) Val() string {
	return this.val
}

func (this *FileParser) GetSection() string {
	return this.section
}

func (this *FileParser) IsNewSection() bool {
	return this.newSection
}

func (this *FileParser) GetRawLine() string {
	if this.scanner != nil && this.scanner.Scan() {
		return strings.TrimSpace(this.scanner.Text())
	}

	return ""
}

func (this *FileParser) IncrementLineNum() {
	this.lineNumber++
}

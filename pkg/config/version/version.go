package version

import (
	"monster/pkg/utils/parsing"
	"strconv"
)

type Version struct {
	X uint16
	Y uint16
	Z uint16
}

func Construct(vars ...uint16) Version {
	v := Version{}
	switch len(vars) {
	case 3:
		v.Z = vars[2]
		fallthrough
	case 2:
		v.Y = vars[1]
		fallthrough
	case 1:
		v.X = vars[0]
	}

	return v
}

func (this *Version) SetFromString(s string) string {
	str := s + "."

	x, str := parsing.PopFirstInt(str, ".")
	this.X = uint16(x)

	strY, str := parsing.PopFirstString(str, ".")
	this.Y = uint16(parsing.ToInt(strY, 0))
	if len(strY) == 1 {
		this.Y = uint16(this.Y * 10)
	}

	strZ, str := parsing.PopFirstString(str, ".")
	this.Z = uint16(parsing.ToInt(strZ, 0))
	if len(strZ) == 1 {
		this.Z = uint16(this.Z * 10)
	}

	return str
}

func (this Version) GetString() string {
	strVersion := strconv.Itoa((int)(this.X)) + "."
	if this.Y >= 100 || this.Y == 0 {
		strVersion += strconv.Itoa((int)(this.Y))
	} else if this.Y >= 10 {
		strVersion += strconv.Itoa((int)(this.Y))
	} else {
		strVersion += "0"
		strVersion += strconv.Itoa((int)(this.Y))
	}

	if this.Z == 0 {
		return strVersion
	}

	strVersion += "."
	if this.Z >= 100 {
		strVersion += strconv.Itoa((int)(this.Z))
	} else if this.Z >= 10 {
		strVersion += strconv.Itoa((int)(this.Z))
	} else {
		strVersion += "0"
		strVersion += strconv.Itoa((int)(this.Z))
	}

	return strVersion
}

func Compare(first Version, second Version) int {
	if first.X == second.X && first.Y == second.Y && first.Z == second.Z {
		return 0
	}

	if first.X > second.X {
		return 1
	} else if first.X < second.X {
		return -1
	}

	if first.Y > second.Y {
		return 1
	} else if first.Y < second.Y {
		return -1
	}

	if first.Z > second.Z {
		return 1
	}

	return -1
}

func CreateVersionReqString(v1 Version, v2 Version) string {
	minVersion := ""
	maxVersion := ""
	ret := ""

	if Compare(v1, MIN) != 0 {
		minVersion = v1.GetString()
	}

	if Compare(v2, MAX) != 0 {
		maxVersion = v2.GetString()
	}

	if minVersion != "" || maxVersion != "" {
		if minVersion == maxVersion {
			ret += minVersion
		} else if minVersion != "" && maxVersion != "" {
			ret += minVersion + "-" + maxVersion
		} else if minVersion != "" {
			//TODO messageEngine
			ret += minVersion + " " + "or newer"
		} else if maxVersion != "" {
			//TODO messageEngine
			ret += maxVersion + " " + "or older"
		}
	}

	return ret
}

package platform

var (
	default_platform = New()
)

func New() *Platform {
	return &Platform{}
}

func (this *Platform) setPaths() {

}

func (this *Platform) FullscreenBypass() bool {
	return false
}

//=============================================
func (this *Platform) SetFullscreen(bool)  {}
func (this *Platform) FSCommit()           {}
func (this *Platform) SetExitEventFilter() {}

package parser

/********************************************************************
 * ZipProvider
 ********************************************************************/

// provide entries from zip archive
type ZipProvider struct {
}

func NewZipProvider(zipname string) (provider *ZipProvider, err error) {
	return
}

func (self *ZipProvider) List() (entries []string) {
	return
}

func (self *ZipProvider) Get(name string) (data []byte, err error) {
	return
}

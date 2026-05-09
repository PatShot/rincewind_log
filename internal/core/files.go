package core

type FileStruct struct {
	LogFile    string
	MainFile   string
	ConfigFile string
	DBFile     string
}

func (f *FileStruct) GetLogFileName() string {
	return f.LogFile
}

func (f *FileStruct) GetMainFileName() string {
	return f.MainFile
}

func (f *FileStruct) GetConfigFileName() string {
	return f.ConfigFile
}

func (f *FileStruct) SetLogFileName(name string) {
	f.LogFile = name
}

func (f *FileStruct) SetMainFileName(name string) {
	f.MainFile = name
}

func (f *FileStruct) SetConfigFileName(name string) {
	f.ConfigFile = name
}

func (f *FileStruct) GetDBFileName() string {
	return f.DBFile
}

func (f *FileStruct) SetDBFileName(name string) {
	f.DBFile = name
}

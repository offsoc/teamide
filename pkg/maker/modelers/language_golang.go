package modelers

import "strings"

type LanguageGolangModel struct {
	ElementNode
	Dir          string `json:"dir,omitempty"`
	ModuleName   string `json:"moduleName,omitempty"`
	GoVersion    string `json:"goVersion,omitempty"`
	CommonPath   string `json:"commonPath,omitempty"`
	CommonPack   string `json:"commonPack,omitempty"`
	ConstantPath string `json:"constantPath,omitempty"`
	ConstantPack string `json:"constantPack,omitempty"`
	ErrorPath    string `json:"errorPath,omitempty"`
	ErrorPack    string `json:"errorPack,omitempty"`
	StructPath   string `json:"structPath,omitempty"`
	StructPack   string `json:"structPack,omitempty"`
	FuncPath     string `json:"funcPath,omitempty"`
	FuncPack     string `json:"funcPack,omitempty"`
}

func (this_ *LanguageGolangModel) GetModuleName() string {
	if this_.ModuleName != "" {
		return this_.ModuleName
	}
	return "app"
}

func (this_ *LanguageGolangModel) GetGoVersion() string {
	if this_.GoVersion != "" {
		return this_.GoVersion
	}
	return "1.18"
}

func (this_ *LanguageGolangModel) GetCommonDir(dir string) string {
	return GetDir(dir, this_.GetCommonPath())
}

func (this_ *LanguageGolangModel) GetCommonPath() string {
	return GetPath(&this_.CommonPath, "common/")
}

func (this_ *LanguageGolangModel) GetCommonPack() string {
	return GetPack(&this_.CommonPack, "common")
}

func (this_ *LanguageGolangModel) GetCommonImport() string {
	return this_.GetPackImport(this_.GetCommonPath(), this_.GetCommonPack())
}

func (this_ *LanguageGolangModel) GetConstantDir(dir string) string {
	return GetDir(dir, this_.GetConstantPath())
}

func (this_ *LanguageGolangModel) GetConstantPath() string {
	return GetPath(&this_.ConstantPath, "constant/")
}

func (this_ *LanguageGolangModel) GetConstantPack() string {
	return GetPack(&this_.ConstantPack, "constant")
}

func (this_ *LanguageGolangModel) GetErrorDir(dir string) string {
	return GetDir(dir, this_.GetErrorPath())
}

func (this_ *LanguageGolangModel) GetErrorPath() string {
	return GetPath(&this_.ErrorPath, "exception/")
}

func (this_ *LanguageGolangModel) GetErrorPack() string {
	return GetPack(&this_.ErrorPack, "exception")
}

func (this_ *LanguageGolangModel) GetStructDir(dir string) string {
	return GetDir(dir, this_.GetStructPath())
}

func (this_ *LanguageGolangModel) GetStructPath() string {
	return GetPath(&this_.StructPath, "bean/")
}

func (this_ *LanguageGolangModel) GetStructPack() string {
	return GetPack(&this_.StructPack, "bean")
}

func (this_ *LanguageGolangModel) GetFuncDir(dir string) string {
	return GetDir(dir, this_.GetFuncPath())
}

func (this_ *LanguageGolangModel) GetFuncPath() string {
	return GetPath(&this_.FuncPath, "tool/")
}

func (this_ *LanguageGolangModel) GetFuncPack() string {
	return GetPack(&this_.FuncPack, "tool")
}

func (this_ *LanguageGolangModel) GetPackImport(path string, pack string) string {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	moduleName := this_.GetModuleName()
	dot := strings.LastIndex(path, "/")
	if dot > 0 {
		moduleName += path[:dot]
	}
	return moduleName + "/" + pack
}

func GetDir(dir string, path string) string {
	return dir + path
}

func GetPath(name *string, defaultPath string) string {
	if *name == "" {
		*name = defaultPath
	} else {
		if !strings.HasSuffix(*name, "/") {
			*name += "/"
		}
	}
	return *name
}

func GetPack(name *string, defaultPack string) string {
	if *name == "" {
		*name = defaultPack
	}

	return *name
}

func init() {
	addDocTemplate(&docTemplate{
		Name:    TypeLanguageGolangName,
		Comment: "语言-Golang",
		Fields: []*docTemplateField{
			{Name: "dir", Comment: "目录"},
			{Name: "moduleName", Comment: "module名称"},
			{Name: "constantPath", Comment: "常量目录路径"},
		},
	})
}
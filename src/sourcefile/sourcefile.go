package sourcefile

import "os"

type SourceFile struct {
	Filename   string
	ByteSource []byte
}

func OpenSource(filename string) (*SourceFile, error) {
	input, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return &SourceFile{filename, input}, nil
}

func (file *SourceFile) Text() string {
	return string(file.ByteSource)
}

func (file *SourceFile) Runes() []rune {
	return []rune(string(file.ByteSource))
}

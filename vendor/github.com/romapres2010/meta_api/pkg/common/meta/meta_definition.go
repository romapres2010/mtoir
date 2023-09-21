package meta

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"encoding/json"

	_err "github.com/romapres2010/meta_api/pkg/common/error"
	_log "github.com/romapres2010/meta_api/pkg/common/logger"
)

// Definition - определение сущности
type Definition struct {
	Name           string           `yaml:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`                      // Имя сущности
	Source         string           `yaml:"source,omitempty" json:"source,omitempty" xml:"source,omitempty"`                // Источник определения сущности REST, File, DB
	Format         string           `yaml:"format,omitempty" json:"format,omitempty" xml:"format,omitempty"`                // Тип определения JSON, YAML, XML
	Json           *json.RawMessage `yaml:"text,omitempty" json:"text,omitempty" xml:"text,omitempty"`                      // Текст определения в читаемом формате
	SourceFileName string           `yaml:"source_file,omitempty" json:"source_file,omitempty" xml:"source_file,omitempty"` // Внешний файл с определением

	Buf []byte `yaml:"-" json:"-" xml:"-"` // Текст определения
}

// LoadFromFile - load from file
func (definition *Definition) LoadFromFile(sourceDir string) error {
	if definition != nil {

		if definition.Source != "File" {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' definition - source must be 'File' '%s'", definition.Name, definition.Source)).PrintfError()
		}

		if definition.SourceFileName == "" {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' definition - empty sourece file name", definition.Name)).PrintfError()
		}

		fileName := sourceDir + definition.SourceFileName

		_log.Info("Loading entity definition from file: Entity, sourceDir, SourceFileName", definition.Name, sourceDir, definition.SourceFileName)

		// Считать информацию о файле
		fileInfo, err := os.Stat(fileName)
		if os.IsNotExist(err) {
			return _err.NewTyped(_err.ERR_CONFIG_FILE_NOT_EXISTS, _err.ERR_UNDEFINED_ID, fileName).PrintfError()
		}

		_log.Debug("Entity definition file exist: FileName, FileInfo", fileName, fileInfo)

		file, err := os.Open(fileName)
		if err != nil {
			return err
		}
		defer func() {
			if file != nil {
				err = file.Close()
				if err != nil {
					_ = _err.WithCauseTyped(_err.ERR_COMMON_ERROR, _err.ERR_UNDEFINED_ID, err, "config.loadYamlConfig -> os.File.Close()").PrintfError()
				}
			}
		}()

		// Read the file into a byte slice
		definition.Buf = make([]byte, fileInfo.Size())
		_, err = bufio.NewReader(file).Read(definition.Buf)
		if err != nil && err != io.EOF {
			return err
		}

		if err != nil {
			return _err.WithCauseTyped(_err.ERR_CONFIG_FILE_LOAD_ERROR, _err.ERR_UNDEFINED_ID, err, fileName).PrintfError()
		}

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if definition != nil {}").PrintfError()
}

func (definition *Definition) check() error {
	if definition != nil {
		if definition.Name == "" {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' definition - empty entity name", definition.Name)).PrintfError()
		}

		if !(definition.Format == "json" || definition.Format == "yaml" || definition.Format == "xml") {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' definition - allowed only 'format'='json', 'yaml', 'xml' '%s'", definition.Name, definition.Format)).PrintfError()
		}

		if !(definition.Source == "REST" || definition.Source == "File" || definition.Source == "DB") {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' definition - allowed only 'source'='REST', 'File', 'DB' '%s'", definition.Name, definition.Source)).PrintfError()
		}

		if len(definition.Buf) == 0 {
			return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' definition - empty definition buffer", definition.Name)).PrintfError()
		}

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if definition != nil {}", []interface{}{definition}).PrintfError()
}

func (meta *Meta) newFromDefinition(definition *Definition) (*Entity, error) {
	if meta != nil && definition != nil {

		entity := meta.newEntity(definition.Name)

		if err := entity.setFromDefinition(definition); err != nil {
			return nil, err
		}

		return entity, nil
	}

	return nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if meta != nil && definition != nil {}", []interface{}{meta, definition}).PrintfError()
}

// reSetFromDefinition - сформировать определение новой сущности через ее определение
func (entity *Entity) resetFromDefinition() (err error) {
	if entity != nil {

		definition := entity.Definition

		// Полная очистка сущности
		entity.clear()

		return entity.setFromDefinition(definition)

	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil {}", []interface{}{entity}).PrintfError()
}

// setFromDefinition - сформировать определение новой сущности через новое определение
func (entity *Entity) setFromDefinition(definition *Definition) (err error) {
	if entity != nil && definition != nil {

		// Только новые сущности
		if entity.isInit {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Entity '%s' - already init - can not be set from definition", entity.Name))
		}

		// проверка корректности
		if err = definition.check(); err != nil {
			return err
		}

		if err = unmarshal(definition.Buf, entity, "setFromDefinition: ", definition.Name, definition.Format); err != nil {
			return err
		}

		if definition.Name != entity.Name {
			return _err.NewTyped(_err.ERR_ERROR, _err.ERR_UNDEFINED_ID, fmt.Sprintf("Definition '%s' - 'definition.name'='%s' does not corespond 'definition.source_file.entity_name'='%s'", definition.Name, definition.Name, entity.Name))
		}

		{ // Подготовим текст definition в формате json для возврата
			var definitionBuf []byte
			definitionBuf, err = json.Marshal(entity)
			if err != nil {
				return _err.WithCauseTyped(_err.ERR_JSON_MARSHAL_ERROR, _err.ERR_UNDEFINED_ID, err, entity).PrintfError()
			}
			definition.Json = (*json.RawMessage)(&definitionBuf)
		} // Подготовим текст definition в формате json для возврата

		// сохраним определение как копию
		definitionCopy := *definition
		entity.Definition = &definitionCopy

		return nil
	}
	return _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, _err.ERR_UNDEFINED_ID, "if entity != nil && definition != nil {}", []interface{}{entity, definition}).PrintfError()
}

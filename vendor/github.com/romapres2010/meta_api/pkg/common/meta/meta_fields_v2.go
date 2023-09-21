package meta

import (
    "google.golang.org/protobuf/proto"
    "google.golang.org/protobuf/reflect/protodesc"
    "google.golang.org/protobuf/reflect/protoregistry"
    "google.golang.org/protobuf/types/descriptorpb"
    "google.golang.org/protobuf/types/dynamicpb"
    "io/ioutil"
    "os/exec"
    "path"
)

type Name string
type FieldNumber int32

type FieldsV2 struct {
    List Fields    // Поля сущности, включая ссылки FK и отношения M:1
    Map  FieldsMap // Поля сущности, включая ссылки FK и отношения M:1

    expr         Exprs             // Формулы вычисления полей
    rules        map[string]string // Правила проверки
    structList   Fields            // Поля сущности, только допустимые к обработке
    structMap    FieldsMap         // Поля сущности, только допустимые к обработке
    dbNameMap    FieldsMap         // Для быстрого поиска и проверки дублей мета модели
    jsonNameMap  FieldsMap         // Для быстрого поиска и проверки дублей мета модели
    xmlNameMap   FieldsMap         // Для быстрого поиска и проверки дублей мета модели
    yamlNameMap  FieldsMap         // Для быстрого поиска и проверки дублей мета модели
    xlsNameMap   FieldsMap         // Для быстрого поиска и проверки дублей мета модели
    exprNameMap  FieldsMap         // Для быстрого поиска и проверки дублей мета модели
    keyFieldsMap FieldsMap         // Поля всех ключей сущности

    //once   sync.Once
    //byName map[Name]*Field        // protected by once
    //byJSON map[string]*Field      // protected by once
    //byText map[string]*Field      // protected by once
    //byNum  map[FieldNumber]*Field // protected by once
}

//func (p *FieldsV2) lazyInit() *FieldsV2 {
//    p.once.Do(func() {
//        if len(p.List) > 0 {
//            p.byName = make(map[Name]*Field, len(p.List))
//            p.byJSON = make(map[string]*Field, len(p.List))
//            p.byText = make(map[string]*Field, len(p.List))
//            p.byNum = make(map[FieldNumber]*Field, len(p.List))
//            for i := range p.List {
//                d := &p.List[i]
//                _ = d
//                //if _, ok := p.byName[d.Name()]; !ok {
//                //    p.byName[d.Name()] = d
//                //}
//                //if _, ok := p.byJSON[d.JSONName()]; !ok {
//                //    p.byJSON[d.JSONName()] = d
//                //}
//                //if _, ok := p.byText[d.TextName()]; !ok {
//                //    p.byText[d.TextName()] = d
//                //}
//                //if _, ok := p.byNum[d.Number()]; !ok {
//                //    p.byNum[d.Number()] = d
//                //}
//            }
//        }
//    })
//    return p
//}

func (p *FieldsV2) Len() int {
    return len(p.List)
}
func (p *FieldsV2) Get(i int) *Field {
    return p.List[i]
}

//func (p *FieldsV2) ByName(s Name) *Field {
//    if d := p.lazyInit().byName[s]; d != nil {
//        return d
//    }
//    return nil
//}
//func (p *FieldsV2) ByJSONName(s string) *Field {
//    if d := p.lazyInit().byJSON[s]; d != nil {
//        return d
//    }
//    return nil
//}
//func (p *FieldsV2) ByTextName(s string) *Field {
//    if d := p.lazyInit().byText[s]; d != nil {
//        return d
//    }
//    return nil
//}
//func (p *FieldsV2) ByNumber(n FieldNumber) *Field {
//    if d := p.lazyInit().byNum[n]; d != nil {
//        return d
//    }
//    return nil
//}

// JSONCamelCase converts a snake_case identifier to a camelCase identifier,
// according to the protobuf JSON specification.
func JSONCamelCase(s string) string {
    var b []byte
    var wasUnderscore bool
    for i := 0; i < len(s); i++ { // proto identifiers are always ASCII
        c := s[i]
        if c != '_' {
            if wasUnderscore && isASCIILower(c) {
                c -= 'a' - 'A' // convert to uppercase
            }
            b = append(b, c)
        }
        wasUnderscore = c == '_'
    }
    return string(b)
}

// JSONSnakeCase converts a camelCase identifier to a snake_case identifier,
// according to the protobuf JSON specification.
func JSONSnakeCase(s string) string {
    var b []byte
    for i := 0; i < len(s); i++ { // proto identifiers are always ASCII
        c := s[i]
        if isASCIIUpper(c) {
            b = append(b, '_')
            c += 'a' - 'A' // convert to lowercase
        }
        b = append(b, c)
    }
    return string(b)
}

func isASCIILower(c byte) bool {
    return 'a' <= c && c <= 'z'
}
func isASCIIUpper(c byte) bool {
    return 'A' <= c && c <= 'Z'
}
func isASCIIDigit(c byte) bool {
    return '0' <= c && c <= '9'
}

func test() error {
    //https://groups.google.com/g/grpc-io/c/fuf7WeFJcNY
    filename := "test1"
    src_dir := "."
    tmp_file := filename + "tmp.pb"
    cmd := exec.Command("protoc", "--descriptor_set_out="+tmp_file, "-I"+src_dir, path.Join(src_dir, filename))
    _ = cmd

    protoFile, err := ioutil.ReadFile(tmp_file)
    pb_set := new(descriptorpb.FileDescriptorSet)
    if err := proto.Unmarshal(protoFile, pb_set); err != nil {
        return err
    }
    pb := pb_set.GetFile()[0]
    fd, err := protodesc.NewFile(pb, protoregistry.GlobalFiles)
    if err = protoregistry.GlobalFiles.RegisterFile(fd); err != nil {
        return err
    }
    msgType := dynamicpb.NewMessageType(fd.Messages().Get(0))
    err = protoregistry.GlobalTypes.RegisterMessage(msgType)

    return nil
}

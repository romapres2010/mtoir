package apiservice

import (
    "context"
    "fmt"
    "os"
    "path"
    "strings"
    "time"

    "io/ioutil"
    "os/exec"
    "path/filepath"

    "google.golang.org/protobuf/proto"
    "google.golang.org/protobuf/reflect/protodesc"
    "google.golang.org/protobuf/reflect/protoreflect"
    "google.golang.org/protobuf/reflect/protoregistry"
    "google.golang.org/protobuf/types/descriptorpb"
    "google.golang.org/protobuf/types/dynamicpb"

    "gopkg.in/src-d/proteus.v1/protobuf"

    "github.com/jhump/protoreflect/desc/protoparse"

    _ctx "github.com/romapres2010/meta_api/pkg/common/ctx"
    _err "github.com/romapres2010/meta_api/pkg/common/error"
    _log "github.com/romapres2010/meta_api/pkg/common/logger"
    _meta "github.com/romapres2010/meta_api/pkg/common/meta"
)

// GetEntityMeta извлечь метаданные entity
func (s *Service) GetEntityMeta(ctx context.Context, entityName, format string) (exists bool, outBuf []byte, myerr error) {

    var requestID = _ctx.FromContextHTTPRequestID(ctx) // RequestID передается через context
    var err error
    var tic = time.Now()
    var entity *_meta.Entity

    exists = false

    _log.Debug("START: requestID, entityName", requestID, entityName)

    if s.storageMap == nil {
        return exists, nil, _err.NewTyped(_err.ERR_INCORRECT_CALL_ERROR, requestID, "Empty API Service", []interface{}{s.storageMap}).PrintfError()
    }

    entity = s.GetEntity(entityName)
    if err != nil {
        return exists, nil, err
    }
    if entity == nil {
        return false, nil, _err.NewTyped(_err.ERR_ERROR, requestID, fmt.Sprintf("Does not found Entity Meta with 'entity_name'='%s'", entityName))
    }

    if format == "proto" {
        tmpPath, err := os.MkdirTemp("", "entity_meta")
        if err != nil {
            return false, nil, err
        }

        msg := &protobuf.Message{}
        {
            msg.Docs = []string{strings.Join([]string{entity.Alias.FullName, entity.Alias.Description}, " ")}
            msg.Name = entity.Name

            for i, field := range entity.Fields {
                if !field.System && field.Reference() == nil {
                    f := &protobuf.Field{
                        Docs:     []string{strings.Join([]string{field.Alias.FullName, field.Alias.Description}, " ")},
                        Name:     field.Name,
                        Pos:      i + 1,
                        Repeated: false,
                        Type:     protobuf.NewBasic("string"),
                        Options:  nil,
                    }
                    msg.Fields = append(msg.Fields, f)
                }
            }
        }

        pkg := &protobuf.Package{
            Name:     "meta." + entity.Name,
            Imports:  []string{"google/protobuf/timestamp.proto"},
            Messages: []*protobuf.Message{msg},
            //Enums:    []*protobuf.Enum{mockEnum},
            //Options:  protobuf.Options{"foo": protobuf.NewLiteralValue("true")},
            //RPCs:     mockRpcs,
        }

        generator := protobuf.NewGenerator(tmpPath)
        err = generator.Generate(pkg)
        if err != nil {
            return false, nil, _err.WithCauseTyped(_err.ERR_COMMON_ERROR, _err.ERR_UNDEFINED_ID, err, "protobuf.NewGenerator(tmpPath)")
        }

        outBuf, err = os.ReadFile(filepath.Join(tmpPath, "generated.proto"))
        if err != nil {
            return false, nil, err
        }

        fileDescriptors, err := protoparse.Parser{ImportPaths: []string{tmpPath}}.ParseFiles("generated.proto")
        if err != nil {
            return false, nil, _err.WithCauseTyped(_err.ERR_COMMON_ERROR, _err.ERR_UNDEFINED_ID, err, "protoparse.Parser{ImportPaths: []string{tmpPath}}.ParseFiles(\"generated.proto\")")
        }

        // Initialize the File descriptor object
        fd, err := protodesc.NewFile(fileDescriptors[0].AsFileDescriptorProto(), protoregistry.GlobalFiles)
        if err != nil {
            return false, nil, err
        }

        // and finally register it.
        err = protoregistry.GlobalFiles.RegisterFile(fd)
        if err != nil {
            return false, nil, err
        }

        mt := dynamicpb.NewMessage(fileDescriptors[0].GetMessageTypes()[0].UnwrapMessage())

        fd0 := mt.Descriptor().Fields().Get(0)
        mt.Set(fd0, protoreflect.ValueOf("test0"))

        fd1 := mt.Descriptor().Fields().Get(1)
        mt.Set(fd1, protoreflect.ValueOf("test1"))

        err = os.RemoveAll(tmpPath)
        if err != nil {
            return false, nil, err
        }

        // сформируем ответ
        outBuf, err = proto.Marshal(mt)
        if err != nil {
            return false, nil, err
        }

    } else {
        // сформируем ответ
        outBuf, err = s.marshal(requestID, entity, "GetEntityMeta", entityName, format)
        if err != nil {
            return false, nil, err
        }
    }

    _log.Debug("SUCCESS: requestID, entityName, duration", requestID, entityName, time.Now().Sub(tic))

    return true, outBuf, nil
}

func registerProtoFile(src_dir string, filename string) error {
    // First, convert the .proto file to a file descriptor set
    tmp_file := filename + "tmp.pb"
    cmd := exec.Command("protoc",
        "--descriptor_set_out="+tmp_file,
        "-I"+src_dir+path.Join(src_dir, filename))

    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    err := cmd.Run()
    if err != nil {
        return err
    }

    defer os.Remove(tmp_file)

    // Now load that temporary file as a file descriptor set protobuf
    protoFile, err := ioutil.ReadFile(tmp_file)
    if err != nil {
        return err
    }

    pb_set := new(descriptorpb.FileDescriptorSet)
    if err := proto.Unmarshal(protoFile, pb_set); err != nil {
        return err
    }

    // We know protoc was invoked with a single .proto file
    pb := pb_set.GetFile()[0]

    // Initialize the File descriptor object
    fd, err := protodesc.NewFile(pb, protoregistry.GlobalFiles)
    if err != nil {
        return err
    }

    // and finally register it.
    return protoregistry.GlobalFiles.RegisterFile(fd)
}

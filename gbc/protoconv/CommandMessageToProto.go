package protoconv

import (
    "fmt"

    "github.com/dualface/go-gbc/gbc/impl"
    "github.com/golang/protobuf/proto"
)

const (
    mainCmdIdMask = 0xffff
    subCmdIdMask  = 0xffff
)

type (
    ProtoMessageCreator func() proto.Message
)

var registry = map[int]ProtoMessageCreator{}

func RegisterCommandMessageToProto(mainCmdId int, subCmdId int, c ProtoMessageCreator) error {
    key := genKey(mainCmdId, subCmdId)

    _, ok := registry[key]
    if ok {
        return fmt.Errorf("command %d:%d already exits", mainCmdId, subCmdId)
    }

    registry[key] = c
    return nil
}

func UnmarshalCommandMessageToProto(msg *impl.CommandMessage) (proto.Message, error) {
    key := genKey(msg.MainCmdId(), msg.SubCmdId())

    c, ok := registry[key]
    if !ok {
        return nil, fmt.Errorf("not found registered command %d:%d", msg.MainCmdId(), msg.SubCmdId())
    }

    pb := c()
    err := proto.Unmarshal(msg.DataBytes(), pb)
    return pb, err
}

func genKey(mainCmdId int, subCmdId int) int {
    mainCmdId &= mainCmdIdMask
    subCmdId &= subCmdIdMask
    return mainCmdId<<16 + subCmdId
}

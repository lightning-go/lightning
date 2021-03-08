module lightning

go 1.14

require (
	github.com/coreos/etcd v3.3.25+incompatible // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.4.2
	github.com/google/uuid v1.1.2 // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/jinzhu/gorm v1.9.12
	github.com/json-iterator/go v1.1.10
	github.com/lestrrat-go/file-rotatelogs v2.3.0+incompatible
	github.com/lestrrat-go/strftime v1.0.1 // indirect
	github.com/lightning-go/lightning v0.0.0-20200330075514-e3701f4ba5bf
	github.com/pkg/errors v0.9.1
	github.com/ramya-rao-a/go-outline v0.0.0-20200117021646-2a048b4510eb // indirect
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/rogpeppe/godef v1.1.2 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.6.0
	go.uber.org/zap v1.16.0 // indirect
	golang.org/x/tools/gopls v0.4.3 // indirect
	google.golang.org/grpc v1.26.0
)

replace github.com/lightning-go/lightning => ../lightning

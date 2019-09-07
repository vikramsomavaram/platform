module github.com/tribehq/platform

go 1.12

require (
	cloud.google.com/go v0.43.0
	contrib.go.opencensus.io/exporter/stackdriver v0.12.2
	github.com/99designs/gqlgen v0.9.3
	github.com/99designs/gqlgen-contrib v0.0.0-20190222015228-c654377d611c
	github.com/agnivade/levenshtein v1.0.2 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/foolin/echo-template v0.0.0-20190415034849-543a88245eec
	github.com/go-redis/redis v0.0.0-20190813142431-c5c4ad6a4cae
	github.com/go-stack/stack v1.8.0
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/uuid v1.1.1
	github.com/gorilla/context v1.1.1
	github.com/gorilla/sessions v1.2.0
	github.com/gorilla/websocket v1.4.0
	github.com/jinzhu/copier v0.0.0-20180308034124-7e38e58719c3
	github.com/joho/godotenv v1.3.0
	github.com/jordan-wright/email v0.0.0-20190218024454-3ea4d25e7cf8
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/echo-contrib v0.6.0
	github.com/labstack/echo/v4 v4.1.8
	github.com/labstack/gommon v0.2.9
	github.com/mikespook/gorbac v2.1.0+incompatible
	github.com/nicksnyder/go-i18n/v2 v2.0.2
	github.com/osteele/liquid v1.2.4
	github.com/osteele/tuesday v1.0.3 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/common v0.2.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spacemonkeygo/openssl v0.0.0-20181017203307-c2dcc5cca94a
	github.com/stretchr/testify v1.3.0
	github.com/stripe/stripe-go v60.14.0+incompatible
	github.com/tidwall/pretty v0.0.0-20190325153808-1166b9ac2b65 // indirect
	github.com/vektah/gqlparser v1.1.2
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c // indirect
	github.com/xdg/stringprep v1.0.0 // indirect
	go.mongodb.org/mongo-driver v1.1.0
	go.opencensus.io v0.22.0
	golang.org/x/crypto v0.0.0-20190605123033-f99c8df09eb5
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sys v0.0.0-20190801041406-cbf593c0f2f3 // indirect
	golang.org/x/text v0.3.2
	google.golang.org/genproto v0.0.0-20190716160619-c506a9f90610
)

replace github.com/golang/lint => golang.org/x/lint v0.0.0-20190409202823-959b441ac422

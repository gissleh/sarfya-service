package main

import (
	"flag"
	"github.com/aws/aws-lambda-go/lambda"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/gissleh/sarfya"
	"github.com/gissleh/sarfya-service/adapters/fwewdictionary"
	"github.com/gissleh/sarfya-service/adapters/jsonemphasisdata"
	"github.com/gissleh/sarfya-service/adapters/templfrontend"
	"github.com/gissleh/sarfya-service/adapters/webapi"
	"github.com/gissleh/sarfya/adapters/jsonstorage"
	"github.com/gissleh/sarfya/adapters/placeholderdictionary"
	"github.com/gissleh/sarfya/sarfyaservice"
	"log"
)

var flagExampleFile = flag.String("example-file", "./data-compiled.json", "File containing example data.")
var flagEmphasisFile = flag.String("emphasis-data", "./stress-data.json", "File containing stress data.")

func main() {
	flag.Parse()

	dict := sarfya.CombinedDictionary{
		sarfya.WithDerivedPoS(fwewdictionary.Global()),
		placeholderdictionary.New(),
	}

	storage, err := jsonstorage.Open(*flagExampleFile, true)
	if err != nil {
		log.Fatalln("Failed to open json storage:", err)
		return
	}

	emphasisStorage, err := jsonemphasisdata.Load(*flagEmphasisFile)
	if err != nil {
		log.Fatalln("Failed to load emphasis:", err)
		return
	}

	svc := &sarfyaservice.Service{Dictionary: dict, Storage: storage, ReadOnly: true}
	api := webapi.SetupWithoutListener()

	webapi.Utils(api.Group("/api/utils"), dict)
	webapi.Examples(api.Group("/api/examples"), svc, emphasisStorage)
	templfrontend.Endpoints(api.Group(""), svc, emphasisStorage)

	lambda.Start(echoadapter.New(api).ProxyWithContext)
}

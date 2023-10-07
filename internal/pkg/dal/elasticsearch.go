package dal

import (
	"fmt"

	"github.com/olivere/elastic/v7"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/conf"
)

var (
	esc *elastic.Client
)

func MustInitElasticSearch() {
	var err error
	esc, err = elastic.NewClient(
		elastic.SetURL(conf.Conf.ES.Host),
		elastic.SetBasicAuth(conf.Conf.ES.Username, conf.Conf.ES.Password),
		elastic.SetSniff(false))

	if err != nil {
		panic(err)
	}

	fmt.Println("es initialized")
}

func GetES() *elastic.Client {
	return esc
}

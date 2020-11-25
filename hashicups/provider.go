package hashicups

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{},
		DataSourcesMap: map[string]*schema.Resource{
			//когда мы определили источник данных, добавляем его в карту источников
			//атрибут принимает имя источники и схему *schema.Resource определенную в data_source_coffee.go
			"hashicups_coffees": dataSourceCoffees(),
		},
	}
}

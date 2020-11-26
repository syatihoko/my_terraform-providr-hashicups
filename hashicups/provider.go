package hashicups

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	//Дополнительные библиотеки необходимые providerConfigure
	"context"
	"github.com/hashicorp-demoapp/hashicups-client-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

//// Provider - без авторизации
//func Provider() *schema.Provider {
//	return &schema.Provider{
//		ResourcesMap: map[string]*schema.Resource{},
//		DataSourcesMap: map[string]*schema.Resource{
//			//когда мы определили источник данных, добавляем его в карту источников
//			//атрибут принимает имя источники и схему *schema.Resource определенную в data_source_coffee.go
//			"hashicups_coffees": dataSourceCoffees(),
//		},
//	}
//}

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		//Определяем дополнително схему поставщика
		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				//используются переменные окружения для значений по умолчанию
				DefaultFunc: schema.EnvDefaultFunc("HASHICUPS_USERNAME", nil),
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("HASHICUPS_PASSWORD", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"hashicups_order": resourceOrder(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			//когда мы определили источник данных, добавляем его в карту источников
			//атрибут принимает имя источники и схему *schema.Resource определенную в data_source_coffee.go
			"hashicups_coffees": dataSourceCoffees(),
			//Источник данных реализованный в сложном чтении
			"hashicups_order": dataSourceOrder(),
		},
		//Определяем дополнително
		ConfigureContextFunc: providerConfigure,
	}
}

//Эта функция извлекает имя пользователя и пароль из схемы провайдера для аутент. и настройки провайдера
//она возваращает interface{} и diag.Diagnostics type - может возвращать несколько errors\warnings для Terraform
func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	username := d.Get("username").(string)
	password := d.Get("password").(string)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	//этот блок для Warnings
	//Сообщения диагностики содержат доп. инф. для отлладки, включая инф. в какой функции возникло предупреждение
	//Появляется вместе с ошибкой для доп. инфо.
	//diags = append(diags, diag.Diagnostic{
	//	Severity: diag.Warning,
	//	Summary:  "Warning Message Summary",
	//	Detail:   "This is the detailed warning message from providerConfigure",
	//})

	//Вернув клиент API HashiCups провайдер сможет получить доступ к API в качестве meta входного параметра
	if (username != "") && (password != "") {
		c, err := hashicups.NewClient(nil, &username, &password)
		if err != nil {
			//diag.FromErr - отбрасывает стандартные ошибки Go, оставляя только ошибки провайдера
			//return nil, diag.FromErr(err) // Ниже расширяем код, чтобы позволить возвращать несколько ошибок
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error, //бывает ещё уровень diag.Warning
				Summary:  "Unable to create HashiCups client",
				Detail:   "Unable to auth user for authenticated HashiCups client",
			})

			return nil, diags
		}

		return c, diags
	}

	c, err := hashicups.NewClient(nil, nil, nil)
	if err != nil {
		//return nil, diag.FromErr(err)   //Смотри выше про расширение функции
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create HashiCups client",
			Detail:   "Unable to auth user for authenticated HashiCups client",
		})

		return nil, diags

	}

	return c, diags
}

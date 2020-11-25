package hashicups

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCoffees() *schema.Resource {
	//Функция возвращает схему ресурса
	return &schema.Resource{
		//Посколько ДатаРесурсы только читаются, определен только контекст чтения
		ReadContext: dataSourceCoffeesRead,
		//Определана схема ресурса
		//Все ресурсы Terraform должны иметь схему, это позволит сопоставить ответ JSON со схемой
		//  /coffees endpoint returns an array of coffees.
		//посмотрим вывод и напишим под него схему  curl localhost:19090/coffees | jq
		Schema: map[string]*schema.Schema{
			// *schema.Schema - указатель на структуру schema.Schema
			"coffees": &schema.Schema{
				//Тип список значений - schema.Resource
				Type: schema.TypeList,
				//Computed - значение не мы задаем, оно вычисляется при создании ресурса.
				Computed: true,
				// & - означает, получить указатель на данную структуру.
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"teaser": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"price": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"image": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"ingredients": &schema.Schema{
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ingredient_id": &schema.Schema{
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceCoffeesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := &http.Client{Timeout: 10 * time.Second}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Создаем запрос GET rto localhost:19090/coffees
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/coffees", "http://localhost:19090"), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	// Сохраняем ответ в "r"
	r, err := client.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	//defer - отложенный вызов фун-ции, чтобы предыдущие вызовы успели завершить свою работу
	defer r.Body.Close()

	// make - создание среза из карты строк 0 длины
	coffees := make([]map[string]interface{}, 0)
	// Декодируем ответ r в экземпляр coffees
	err = json.NewDecoder(r.Body).Decode(&coffees)
	if err != nil {
		return diag.FromErr(err)
	}

	// d *schema.ResourceData
	// функция Set устанавливает тело ответа (список объектов кофе) схеме в Терроформ, сопоставляя каждому элементу
	// схемы свое значение
	if err := d.Set("coffees", coffees); err != nil {
		return diag.FromErr(err)
	}

	// always run
	// используем SetId для установки идентификатора ресурса
	// Наличие идентификатора скажет об успешном создании ресурса, он используется для повторного чтения ресурса
	// у ресурса нет уникального идент.,
	//идентификатор привязан к тек. времени, поэтому будет обновляться при обновлении ресурса
	//!! По хорошему нужно делать проверку сущ. ресурса и если его нет удалять из стайта.
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	// возвращаем ошибки\предупреждения в Терраформ
	return diags
}

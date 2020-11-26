package hashicups

import (
	"context"
	"time"

	//strconv.Itoa - преобразует числа в строки
	hc "github.com/hashicorp-demoapp/hashicups-client-go" //Клиентская библиотека HashiCups API
	"strconv"                                             //для использования функции преобразования идентификитора строки

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

//Для создания заказа мы отправляем запрос POST в /orders со спискок, содержащим объект кофе и его кол-во:
//curl -X POST -H "Authorization: ${HASHICUPS_TOKEN}" localhost:19090/orders -d '[{"coffee": { "id":1 }, "quantity":4}, {"coffee": { "id":3 }, "quantity":3}]'
//Ответ:
//{"id":4,"items":[{"coffee":{"id":1,"name":"Packer Spiced Latte","teaser":"Packed with goodness to spice up your images","description":"","price":350,"image":"/packer.png","ingredients":null},"quantity":4},{"coffee":{"id":3,"name":"Nomadicano","teaser":"Drink one today and you will want to schedule another","description":"","price":150,"image":"/nomad.png","ingredients":null},"quantity":3}]}
func resourceOrder() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOrderCreate,
		ReadContext:   resourceOrderRead,
		UpdateContext: resourceOrderUpdate,
		DeleteContext: resourceOrderDelete,
		//Схема для сопоставления с завросом выше
		//Здесь нет ИД в отличии от Дата, так как ИД-заказа формируется после создания заказа
		Schema: map[string]*schema.Schema{
			"last_updated": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"items": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"coffee": &schema.Schema{
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": &schema.Schema{
										Type:     schema.TypeInt,
										Required: true,
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
								},
							},
						},
						"quantity": &schema.Schema{
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
		},
	}
}

//m- параметр, содержит клиент HashiCups API определеный в ConfigureContextFunc
func resourceOrderCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hc.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	//d.Get- получаем атрибуты заказа, они размещены в "items" и определены в схеме, и преобразует их в тип  []interface{}
	items := d.Get("items").([]interface{})
	//затем функция преобразует элементы items в срез\slice, требуемой структуры для созд.  заказа
	ois := []hc.OrderItem{}

	for _, item := range items {
		i := item.(map[string]interface{})

		co := i["coffee"].([]interface{})[0]
		coffee := co.(map[string]interface{})

		oi := hc.OrderItem{
			Coffee: hc.Coffee{
				ID: coffee["id"].(int),
			},
			Quantity: i["quantity"].(int),
		}

		ois = append(ois, oi)
	}

	o, err := c.CreateOrder(ois)
	if err != nil {
		return diag.FromErr(err)
	}

	//Устанавливаем ИД ресурса = ИД заказа
	d.SetId(strconv.Itoa(o.ID))

	//Вызываем фун-ю чтения в конце, Это заполнит state терраформ, после создания ресурса
	resourceOrderRead(ctx, d, m)

	return diags
}

func resourceOrderRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hc.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	orderID := d.Id()

	order, err := c.GetOrder(orderID)
	if err != nil {
		return diag.FromErr(err)
	}

	orderItems := flattenOrderItems(&order.Items)
	if err := d.Set("items", orderItems); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceOrderUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*hc.Client)

	orderID := d.Id()

	//Если есть расхождение, то функция обновит заказ с новой конф-ей
	if d.HasChange("items") {
		items := d.Get("items").([]interface{})
		ois := []hc.OrderItem{}

		for _, item := range items {
			i := item.(map[string]interface{})

			co := i["coffee"].([]interface{})[0]
			coffee := co.(map[string]interface{})

			oi := hc.OrderItem{
				Coffee: hc.Coffee{
					ID: coffee["id"].(int),
				},
				Quantity: i["quantity"].(int),
			}
			ois = append(ois, oi)
		}

		_, err := c.UpdateOrder(orderID, ois)
		if err != nil {
			return diag.FromErr(err)
		}
		//Обновляем атрибут до текущей отметки времени
		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceOrderRead(ctx, d, m)
}

func resourceOrderDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}

//Сглаживание, получение из среза карты
func flattenOrderItems(orderItems *[]hc.OrderItem) []interface{} {
	if orderItems != nil {
		ois := make([]interface{}, len(*orderItems), len(*orderItems))

		for i, orderItem := range *orderItems {
			oi := make(map[string]interface{})

			oi["coffee"] = flattenCoffee(orderItem.Coffee)
			oi["quantity"] = orderItem.Quantity
			ois[i] = oi
		}

		return ois
	}

	return make([]interface{}, 0)
}

//Сглаживание, делаем список с одним элементом, как в нужной схеме кофе.
func flattenCoffee(coffee hc.Coffee) []interface{} {
	c := make(map[string]interface{})
	c["id"] = coffee.ID
	c["name"] = coffee.Name
	c["teaser"] = coffee.Teaser
	c["description"] = coffee.Description
	c["price"] = coffee.Price
	c["image"] = coffee.Image

	return []interface{}{c}
}

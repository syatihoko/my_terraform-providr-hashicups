package hashicups

import (
	"context"
	"strconv"

	hc "github.com/hashicorp-demoapp/hashicups-client-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceOrder() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOrderRead,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"items": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"coffee_id": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"coffee_name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"coffee_teaser": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"coffee_description": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"coffee_price": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"coffee_image": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"quantity": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

//Реализация сложного чтения
func dataSourceOrderRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//m-  входной параметр(мета) содержит клиент API HashiCups, установленный ConfigureContextFunc выше
	//без аутентификации эта функция не сработает
	c := m.(*hc.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	//
	orderID := strconv.Itoa(d.Get("id").(int))

	//GetOrder - получить объект заказа orderID
	order, err := c.GetOrder(orderID)
	if err != nil {
		return diag.FromErr(err)
	}

	// нужно сделать "сплющивание", что бы сопоставить схеме
	//Есть функция flatten - из списка списков получить один список.
	//flattenOrderItemsData - Функция чтобы "сгладить" ответ и точно сопоставить его с имеющейся схемой
	orderItems := flattenOrderItemsData(&order.Items)
	if err := d.Set("items", orderItems); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(orderID)

	return diags
}

//Функция чтобы "сгладить" ответ и точно сопоставить его с имеющейся схемой
//The flattenOrderItemsData function takes an *[]hc.OrderItem as orderItems.
//If orderItems is not nil, it will iterate through the slice and map its values into a map[string]interface{}.
//Notice how the function assigns the coffee attributes directly
//to its corresponding flattened attribute (orderItem.Coffee.ID -> coffee_id).
func flattenOrderItemsData(orderItems *[]hc.OrderItem) []interface{} {
	if orderItems != nil {
		// make - создание среза, длина, емкость
		ois := make([]interface{}, len(*orderItems), len(*orderItems))

		for i, orderItem := range *orderItems {
			oi := make(map[string]interface{})

			oi["coffee_id"] = orderItem.Coffee.ID
			oi["coffee_name"] = orderItem.Coffee.Name
			oi["coffee_teaser"] = orderItem.Coffee.Teaser
			oi["coffee_description"] = orderItem.Coffee.Description
			oi["coffee_price"] = orderItem.Coffee.Price
			oi["coffee_image"] = orderItem.Coffee.Image
			oi["quantity"] = orderItem.Quantity

			ois[i] = oi
		}

		return ois
	}

	return make([]interface{}, 0)
}

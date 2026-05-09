package domain

type Center struct {
	ID                      string  `json:"id" dynamodbav:"id"`
	Name                    string  `json:"name" dynamodbav:"name"`
	PhoneNumber             string  `json:"phone_number" dynamodbav:"phone_number"`
	Address                 string  `json:"address" dynamodbav:"address"`
	WhatsappBusinessPhoneID *string `json:"whatsapp_business_phone_id" dynamodbav:"whatsapp_business_phone_id"`
}

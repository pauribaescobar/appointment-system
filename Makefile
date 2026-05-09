PROFILE = personal
REGION = eu-west-3
ACCOUNT_ID = $(shell aws sts get-caller-identity --profile $(PROFILE) --query Account --output text)

CONFIRMATION_SENDER_FUNCTION = appointment-confirmation-sender
CONFIRMATION_SENDER_SRC = cmd/appointment-confirmation-sender/main.go

WEBHOOK_FUNCTION = whatsapp-webhook
WEBHOOK_SRC = cmd/whatsapp-webhook/main.go
LAMBDA_ROLE = arn:aws:iam::820580909625:role/confirmation-service-role

include .env.lambda

LAMBDA_ROLE_NAME = confirmation-service-role
APPOINTMENTS_TABLE_ARN = arn:aws:dynamodb:$(REGION):820580909625:table/appointments
CENTERS_TABLE_ARN = arn:aws:dynamodb:$(REGION):820580909625:table/centers
FLOWWW_OPERATIONS_ARN = arn:aws:lambda:$(REGION):820580909625:function:flowww-operations
FLOWWW_OPERATIONS_FUNCTION = flowww-operations
LAMBDA_TIMEOUT = 600

.PHONY: build-confirmation-sender deploy-confirmation-sender configure-confirmation-sender deploy-confirmation-sender-full \
	build-webhook deploy-webhook configure-webhook deploy-webhook-full create-webhook \
	setup-api-gateway setup-message-id-index setup-iam-permissions setup-lambda-timeouts

# --- Confirmation Sender ---

build-confirmation-sender:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap $(CONFIRMATION_SENDER_SRC)
	zip $(CONFIRMATION_SENDER_FUNCTION).zip bootstrap
	rm bootstrap

deploy-confirmation-sender: build-confirmation-sender
	aws lambda update-function-code \
		--function-name $(CONFIRMATION_SENDER_FUNCTION) \
		--zip-file fileb://$(CONFIRMATION_SENDER_FUNCTION).zip \
		--profile $(PROFILE) --region $(REGION)
	rm $(CONFIRMATION_SENDER_FUNCTION).zip

configure-confirmation-sender:
	aws lambda update-function-configuration \
		--function-name $(CONFIRMATION_SENDER_FUNCTION) \
		--environment "Variables={DYNAMODB_APPOINTMENTS_TABLE=$(DYNAMODB_APPOINTMENTS_TABLE),DYNAMODB_CENTERS_TABLE=$(DYNAMODB_CENTERS_TABLE),ENV=$(ENV),PHONE_HASH_SALT=$(PHONE_HASH_SALT),LOG_LEVEL=$(LOG_LEVEL),WHATSAPP_AUTH_TOKEN=$(WHATSAPP_AUTH_TOKEN)}" \
		--profile $(PROFILE) --region $(REGION)

deploy-confirmation-sender-full: configure-confirmation-sender deploy-confirmation-sender

# --- Webhook ---

build-webhook:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap $(WEBHOOK_SRC)
	zip $(WEBHOOK_FUNCTION).zip bootstrap
	rm bootstrap

deploy-webhook: build-webhook
	aws lambda update-function-code \
		--function-name $(WEBHOOK_FUNCTION) \
		--zip-file fileb://$(WEBHOOK_FUNCTION).zip \
		--profile $(PROFILE) --region $(REGION)
	rm $(WEBHOOK_FUNCTION).zip

configure-webhook:
	aws lambda update-function-configuration \
		--function-name $(WEBHOOK_FUNCTION) \
		--environment "Variables={DYNAMODB_APPOINTMENTS_TABLE=$(DYNAMODB_APPOINTMENTS_TABLE),DYNAMODB_CENTERS_TABLE=$(DYNAMODB_CENTERS_TABLE),ENV=$(ENV),PHONE_HASH_SALT=$(PHONE_HASH_SALT),LOG_LEVEL=$(LOG_LEVEL),WHATSAPP_AUTH_TOKEN=$(WHATSAPP_AUTH_TOKEN),WEBHOOK_VERIFY_TOKEN=$(WEBHOOK_VERIFY_TOKEN)}" \
		--profile $(PROFILE) --region $(REGION)

deploy-webhook-full: configure-webhook deploy-webhook

create-webhook: build-webhook
	aws lambda create-function \
		--function-name $(WEBHOOK_FUNCTION) \
		--runtime provided.al2023 \
		--architectures arm64 \
		--handler bootstrap \
		--role $(LAMBDA_ROLE) \
		--zip-file fileb://$(WEBHOOK_FUNCTION).zip \
		--profile $(PROFILE) --region $(REGION) \
	|| echo "Function already exists, skipping create"
	rm -f $(WEBHOOK_FUNCTION).zip

# --- One-time Infrastructure Setup ---

setup-api-gateway:
	@echo "Creating HTTP API..."
	$(eval API_ID := $(shell aws apigatewayv2 create-api \
		--name "appointment-webhook-api" \
		--protocol-type HTTP \
		--profile $(PROFILE) --region $(REGION) \
		--query ApiId --output text))
	@echo "API created: $(API_ID)"
	@echo "Creating Lambda integration..."
	$(eval INTEGRATION_ID := $(shell aws apigatewayv2 create-integration \
		--api-id $(API_ID) \
		--integration-type AWS_PROXY \
		--integration-uri arn:aws:lambda:$(REGION):$(ACCOUNT_ID):function:$(WEBHOOK_FUNCTION) \
		--payload-format-version 2.0 \
		--profile $(PROFILE) --region $(REGION) \
		--query IntegrationId --output text))
	@echo "Integration created: $(INTEGRATION_ID)"
	aws apigatewayv2 create-route \
		--api-id $(API_ID) \
		--route-key "GET /webhook" \
		--target integrations/$(INTEGRATION_ID) \
		--profile $(PROFILE) --region $(REGION)
	aws apigatewayv2 create-route \
		--api-id $(API_ID) \
		--route-key "POST /webhook" \
		--target integrations/$(INTEGRATION_ID) \
		--profile $(PROFILE) --region $(REGION)
	aws apigatewayv2 create-stage \
		--api-id $(API_ID) \
		--stage-name '$$default' \
		--auto-deploy \
		--profile $(PROFILE) --region $(REGION)
	aws lambda add-permission \
		--function-name $(WEBHOOK_FUNCTION) \
		--statement-id AllowAPIGateway \
		--action lambda:InvokeFunction \
		--principal apigateway.amazonaws.com \
		--source-arn "arn:aws:execute-api:$(REGION):$(ACCOUNT_ID):$(API_ID)/*" \
		--profile $(PROFILE) --region $(REGION)
	@echo "Webhook URL: https://$(API_ID).execute-api.$(REGION).amazonaws.com/webhook"

setup-message-id-index:
	aws dynamodb update-table \
		--table-name $(DYNAMODB_APPOINTMENTS_TABLE) \
		--attribute-definitions AttributeName=message_id,AttributeType=S \
		--global-secondary-index-updates '[{"Create":{"IndexName":"message_id-index","KeySchema":[{"AttributeName":"message_id","KeyType":"HASH"}],"Projection":{"ProjectionType":"ALL"}}}]' \
		--profile $(PROFILE) --region $(REGION)

setup-iam-permissions:
	aws iam put-role-policy \
		--role-name $(LAMBDA_ROLE_NAME) \
		--policy-name AppointmentServiceAccess \
		--policy-document '{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["dynamodb:Query","dynamodb:GetItem","dynamodb:PutItem","dynamodb:UpdateItem","dynamodb:DeleteItem"],"Resource":["$(APPOINTMENTS_TABLE_ARN)","$(APPOINTMENTS_TABLE_ARN)/index/*","$(CENTERS_TABLE_ARN)"]},{"Effect":"Allow","Action":["lambda:InvokeFunction"],"Resource":["$(FLOWWW_OPERATIONS_ARN)"]}]}' \
		--profile $(PROFILE) --region $(REGION)
	@echo "IAM policy applied to $(LAMBDA_ROLE_NAME)"

setup-lambda-timeouts:
	aws lambda update-function-configuration \
		--function-name $(WEBHOOK_FUNCTION) \
		--timeout $(LAMBDA_TIMEOUT) \
		--profile $(PROFILE) --region $(REGION)
	aws lambda update-function-configuration \
		--function-name $(CONFIRMATION_SENDER_FUNCTION) \
		--timeout $(LAMBDA_TIMEOUT) \
		--profile $(PROFILE) --region $(REGION)
	aws lambda update-function-configuration \
		--function-name $(FLOWWW_OPERATIONS_FUNCTION) \
		--timeout $(LAMBDA_TIMEOUT) \
		--profile $(PROFILE) --region $(REGION)
	@echo "Timeout set to $(LAMBDA_TIMEOUT)s for all Lambda functions"

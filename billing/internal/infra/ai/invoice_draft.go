package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"billing/internal/model"

	"github.com/anthropics/anthropic-sdk-go"
)

const extractionTimeout = 30 * time.Second

const instructions = `Você extrai itens de nota fiscal a partir de um pedido em português.

Devolva um JSON com:
- "type": "IN" quando o pedido descreve entrada de mercadoria (recebi, comprei,
  entrada, do fornecedor) e "OUT" quando descreve saída (vender, enviar, saída,
  para o cliente). Use "OUT" quando o texto não deixar claro.
- "items": um objeto por produto pedido, com:
  - "text": o trecho do pedido que nomeia o produto, sem a quantidade
  - "code": o código do produto do catálogo abaixo, quando você reconhecer qual é
  - "quantity": a quantidade pedida, como número inteiro

Regras:
- Só preencha "code" com um código que esteja no catálogo. Se estiver em dúvida
  entre dois produtos, deixe "code" vazio e preencha só "text".
- Nunca invente produto que o pedido não menciona.
- Não calcule nem informe preços.
- Quando o pedido não disser a quantidade, use 1.`

type InvoiceDraftExtractor struct {
	client anthropic.Client
}

func NewInvoiceDraftExtractor() model.InvoiceDraftExtractor {
	client, configured := newClient()
	if !configured {
		fmt.Println("ANTHROPIC_API_KEY not set: invoice draft extraction disabled")
		return nil
	}

	return &InvoiceDraftExtractor{client: client}
}

func (e *InvoiceDraftExtractor) Extract(
	ctx context.Context,
	prompt string,
	catalog []model.Product,
) (model.InvoiceDraftExtraction, error) {
	ctx, cancel := context.WithTimeout(ctx, extractionTimeout)
	defer cancel()

	message, err := e.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     draftModel,
		MaxTokens: maxTokens,
		System: []anthropic.TextBlockParam{
			{
				Text:         instructions + "\n\n" + catalogTable(catalog),
				CacheControl: anthropic.NewCacheControlEphemeralParam(),
			},
		},
		OutputConfig: anthropic.OutputConfigParam{
			Effort: anthropic.OutputConfigEffortLow,
			Format: anthropic.JSONOutputFormatParam{Schema: extractionSchema},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return model.InvoiceDraftExtraction{}, err
	}

	return decode(message)
}

func decode(message *anthropic.Message) (model.InvoiceDraftExtraction, error) {
	var payload struct {
		Type  string `json:"type"`
		Items []struct {
			Text     string `json:"text"`
			Code     string `json:"code"`
			Quantity int    `json:"quantity"`
		} `json:"items"`
	}

	var text strings.Builder

	for _, block := range message.Content {
		if variant, ok := block.AsAny().(anthropic.TextBlock); ok {
			text.WriteString(variant.Text)
		}
	}

	if text.Len() == 0 {
		return model.InvoiceDraftExtraction{}, errors.New("empty extraction response")
	}

	if err := json.Unmarshal([]byte(text.String()), &payload); err != nil {
		return model.InvoiceDraftExtraction{}, err
	}

	extraction := model.InvoiceDraftExtraction{
		Type:  payload.Type,
		Items: make([]model.ExtractedItem, 0, len(payload.Items)),
	}

	for _, item := range payload.Items {
		extraction.Items = append(extraction.Items, model.ExtractedItem{
			Text:     item.Text,
			Code:     item.Code,
			Quantity: item.Quantity,
		})
	}

	return extraction, nil
}

func catalogTable(catalog []model.Product) string {
	var builder strings.Builder

	builder.WriteString("Catálogo de produtos (código — nome):\n")

	for _, product := range catalog {
		if !product.Active {
			continue
		}

		fmt.Fprintf(&builder, "%s — %s\n", product.Code, product.Name)
	}

	return builder.String()
}

var extractionSchema = map[string]any{
	"type":                 "object",
	"additionalProperties": false,
	"required":             []string{"type", "items"},
	"properties": map[string]any{
		"type": map[string]any{
			"type": "string",
			"enum": []string{model.InvoiceTypeIn, model.InvoiceTypeOut},
		},
		"items": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"required":             []string{"text", "code", "quantity"},
				"properties": map[string]any{
					"text":     map[string]any{"type": "string"},
					"code":     map[string]any{"type": "string"},
					"quantity": map[string]any{"type": "integer"},
				},
			},
		},
	},
}

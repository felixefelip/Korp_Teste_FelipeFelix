package model

import (
	"errors"
	"strings"
	"unicode"
)

const (
	DraftNotFound        = "NOT_FOUND"
	DraftAmbiguous       = "AMBIGUOUS"
	DraftInvalidQuantity = "INVALID_QUANTITY"
)

const minimumTokenMatch = 3

var ErrDraftUnavailable = errors.New("invoice draft extraction unavailable")

type ExtractedItem struct {
	Text     string
	Code     string
	Quantity int
}

type InvoiceDraftExtraction struct {
	Type  string
	Items []ExtractedItem
}

type UnresolvedItem struct {
	Text       string
	Quantity   int
	Reason     string
	Candidates []Product
}

type InvoiceDraft struct {
	Type       string
	Items      []InvoiceItem
	Unresolved []UnresolvedItem
}

func ResolveInvoiceDraft(extraction InvoiceDraftExtraction, catalog []Product) InvoiceDraft {
	draft := InvoiceDraft{
		Type:       draftType(extraction.Type),
		Items:      []InvoiceItem{},
		Unresolved: []UnresolvedItem{},
	}

	active := activeProducts(catalog)

	for _, extracted := range extraction.Items {
		if extracted.Quantity <= 0 {
			draft.Unresolved = append(draft.Unresolved, UnresolvedItem{
				Text:       strings.TrimSpace(extracted.Text),
				Quantity:   extracted.Quantity,
				Reason:     DraftInvalidQuantity,
				Candidates: []Product{},
			})

			continue
		}

		candidates := matchProduct(extracted, active)

		if len(candidates) == 1 {
			draft.Items = append(draft.Items, newDraftItem(candidates[0], extracted.Quantity))
			continue
		}

		reason := DraftNotFound
		if len(candidates) > 1 {
			reason = DraftAmbiguous
		}

		draft.Unresolved = append(draft.Unresolved, UnresolvedItem{
			Text:       strings.TrimSpace(extracted.Text),
			Quantity:   extracted.Quantity,
			Reason:     reason,
			Candidates: candidates,
		})
	}

	return draft
}

func newDraftItem(product Product, quantity int) InvoiceItem {
	return InvoiceItem{
		ProductCode: product.Code,
		ProductName: product.Name,
		Unit:        product.Unit,
		Quantity:    quantity,
		UnitPrice:   product.Price,
		Product:     product,
	}
}

func matchProduct(extracted ExtractedItem, catalog []Product) []Product {
	if product, found := findByCode(extracted.Code, catalog); found {
		return []Product{product}
	}

	tokens := tokenize(extracted.Text)
	text := strings.Join(tokens, " ")

	if product, found := findCodeInText(text, catalog); found {
		return []Product{product}
	}

	for _, product := range catalog {
		if text != "" && text == strings.Join(tokenize(product.Name), " ") {
			return []Product{product}
		}
	}

	if candidates := matchByTokens(tokens, catalog, nameCoveredByText); len(candidates) > 0 {
		return candidates
	}

	return matchByTokens(tokens, catalog, textCoveredByName)
}

func findByCode(code string, catalog []Product) (Product, bool) {
	normalized := normalize(code)
	if normalized == "" {
		return Product{}, false
	}

	for _, product := range catalog {
		if normalize(product.Code) == normalized {
			return product, true
		}
	}

	return Product{}, false
}

func findCodeInText(text string, catalog []Product) (Product, bool) {
	if text == "" {
		return Product{}, false
	}

	padded := " " + text + " "

	for _, product := range catalog {
		code := normalize(product.Code)
		if code != "" && strings.Contains(padded, " "+code+" ") {
			return product, true
		}
	}

	return Product{}, false
}

func matchByTokens(
	tokens []string,
	catalog []Product,
	covers func(name, text []string) bool,
) []Product {
	if len(tokens) == 0 {
		return nil
	}

	matched := []Product{}

	for _, product := range catalog {
		name := tokenize(product.Name)
		if len(name) == 0 {
			continue
		}

		if covers(name, tokens) {
			matched = append(matched, product)
		}
	}

	return matched
}

func nameCoveredByText(name, text []string) bool {
	return covered(name, text)
}

func textCoveredByName(name, text []string) bool {
	return covered(text, name)
}

func covered(needles, haystack []string) bool {
	for _, needle := range needles {
		if !matchesAny(needle, haystack) {
			return false
		}
	}

	return true
}

func matchesAny(token string, tokens []string) bool {
	for _, candidate := range tokens {
		if tokensMatch(token, candidate) {
			return true
		}
	}

	return false
}

func tokensMatch(left, right string) bool {
	if left == right {
		return true
	}

	shorter, longer := left, right
	if len(shorter) > len(longer) {
		shorter, longer = longer, shorter
	}

	return len(shorter) >= minimumTokenMatch && strings.HasPrefix(longer, shorter)
}

func tokenize(text string) []string {
	normalized := normalize(text)
	if normalized == "" {
		return nil
	}

	return strings.Fields(normalized)
}

func normalize(text string) string {
	var builder strings.Builder

	for _, symbol := range strings.ToLower(strings.TrimSpace(text)) {
		switch {
		case unicode.IsLetter(symbol) || unicode.IsDigit(symbol):
			builder.WriteRune(fold(symbol))
		case unicode.IsSpace(symbol):
			builder.WriteRune(' ')
		default:
			builder.WriteRune(' ')
		}
	}

	return strings.Join(strings.Fields(builder.String()), " ")
}

func fold(symbol rune) rune {
	switch symbol {
	case 'á', 'à', 'ã', 'â', 'ä':
		return 'a'
	case 'é', 'è', 'ê', 'ë':
		return 'e'
	case 'í', 'ì', 'î', 'ï':
		return 'i'
	case 'ó', 'ò', 'õ', 'ô', 'ö':
		return 'o'
	case 'ú', 'ù', 'û', 'ü':
		return 'u'
	case 'ç':
		return 'c'
	case 'ñ':
		return 'n'
	default:
		return symbol
	}
}

func activeProducts(catalog []Product) []Product {
	active := make([]Product, 0, len(catalog))

	for _, product := range catalog {
		if product.Active {
			active = append(active, product)
		}
	}

	return active
}

func draftType(value string) string {
	if strings.ToUpper(strings.TrimSpace(value)) == InvoiceTypeIn {
		return InvoiceTypeIn
	}

	return InvoiceTypeOut
}

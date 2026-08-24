package invoice

type draftRequest struct {
	Prompt string `json:"prompt" binding:"required,max=1000"`
}

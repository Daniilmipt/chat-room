package models

type Messages struct {
	Room    string `json:"room"`
	Nick    string `json:"nick"`
	Message string `json:"message"`
}

func (r *Messages) Validate() bool {
	return r.Room != "" && r.Nick != "" && r.Message != ""
}

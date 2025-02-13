package models

type Messages struct {
	Room    string `json:"room"`
	Nick    string `json:"nick"`
	Message []byte `json:"message"`
}

func (r *Messages) Validate() bool {
	return r.Room != "" && r.Nick != "" && len(r.Message) != 0
}

package tickets

import "testing"

func TestBuildTicketChannelName(t *testing.T) {
	tests := []struct {
		name, style, username, interactionID, want string
	}{
		{"number", "number", "Sushi", "12345678", "ticket-5678"},
		{"username", "name", "Sushi", "12345678", "ticket-sushi"},
		{"combined", "name_number", "Sushi", "12345678", "ticket-sushi-5678"},
		{"combined short number", "name_number", "Sushi", "12", "ticket-sushi-12"},
		{"combined normalization", "name_number", "Sushi User", "12345678", "ticket-sushi-user-5678"},
		{"combined missing username", "name_number", "", "12345678", "ticket-5678"},
		{"missing username", "name", "", "12345678", "ticket-5678"},
		{"unknown style", "unknown", "Sushi", "12345678", "ticket-5678"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildTicketChannelName(tt.style, tt.username, tt.interactionID); got != tt.want {
				t.Fatalf("buildTicketChannelName() = %q, want %q", got, tt.want)
			}
		})
	}
}

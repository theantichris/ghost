package conversation

import (
	"slices"
	"testing"
)

func TestNewSession(t *testing.T) {
	tests := []struct {
		name    string
		initial []Message
		want    []Message
	}{
		{
			name: "starts empty",
		},
		{
			name: "preserves initial context",
			initial: []Message{
				{
					Role:    RoleSystem,
					Content: "You are Ghost.",
				},
				{
					Role:    RoleUser,
					Content: "Establish uplink.",
				},
			},
			want: []Message{
				{
					Role:    RoleSystem,
					Content: "You are Ghost.",
				},
				{
					Role:    RoleUser,
					Content: "Establish uplink.",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			initial := slices.Clone(test.initial)
			session := NewSession(initial...)

			if len(initial) > 0 {
				initial[0] = Message{
					Role:    RoleAssistant,
					Content: "mutated outside the session",
				}
			}

			if got := session.Messages(); !slices.Equal(got, test.want) {
				t.Errorf("Messages() = %#v, want %#v", got, test.want)
			}
		})
	}

	t.Run("sessions do not share history", func(t *testing.T) {
		first := NewSession()
		second := NewSession()

		first.AppendUser("first session")

		if got := second.Messages(); len(got) != 0 {
			t.Errorf("second Messages() = %#v, want empty history", got)
		}
	})
}

func TestSessionAppendUser(t *testing.T) {
	tests := []struct {
		name    string
		initial []Message
		content string
		want    []Message
	}{
		{
			name:    "appends to empty history",
			content: "Identify yourself.",
			want: []Message{
				{
					Role:    RoleUser,
					Content: "Identify yourself.",
				},
			},
		},
		{
			name: "appends after initial context",
			initial: []Message{
				{
					Role:    RoleSystem,
					Content: "You are Ghost.",
				},
			},
			content: "Identify yourself.",
			want: []Message{
				{
					Role:    RoleSystem,
					Content: "You are Ghost.",
				},
				{
					Role:    RoleUser,
					Content: "Identify yourself.",
				},
			},
		},
		{
			name:    "preserves content verbatim",
			content: "  signal with surrounding space  ",
			want: []Message{
				{
					Role:    RoleUser,
					Content: "  signal with surrounding space  ",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			session := NewSession(test.initial...)
			session.AppendUser(test.content)

			if got := session.Messages(); !slices.Equal(got, test.want) {
				t.Errorf("Messages() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestSessionAppendAssistant(t *testing.T) {
	tests := []struct {
		name    string
		initial []Message
		content string
		want    []Message
	}{
		{
			name:    "appends to empty history",
			content: "Ghost online.",
			want: []Message{
				{
					Role:    RoleAssistant,
					Content: "Ghost online.",
				},
			},
		},
		{
			name: "appends after user message",
			initial: []Message{
				{
					Role:    RoleUser,
					Content: "Identify yourself.",
				},
			},
			content: "Ghost online.",
			want: []Message{
				{
					Role:    RoleUser,
					Content: "Identify yourself.",
				},
				{
					Role:    RoleAssistant,
					Content: "Ghost online.",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			session := NewSession(test.initial...)
			session.AppendAssistant(test.content)

			if got := session.Messages(); !slices.Equal(got, test.want) {
				t.Errorf("Messages() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestSessionMessages(t *testing.T) {
	tests := []struct {
		name    string
		initial []Message
		build   func(*Session)
		want    []Message
	}{
		{
			name: "returns empty history",
		},
		{
			name: "returns complete history in order",
			initial: []Message{
				{
					Role:    RoleSystem,
					Content: "You are Ghost.",
				},
			},
			build: func(session *Session) {
				session.AppendUser("First signal.")
				session.AppendAssistant("First response.")
				session.AppendUser("Second signal.")
			},
			want: []Message{
				{
					Role:    RoleSystem,
					Content: "You are Ghost.",
				},
				{
					Role:    RoleUser,
					Content: "First signal.",
				},
				{
					Role:    RoleAssistant,
					Content: "First response.",
				},
				{
					Role:    RoleUser,
					Content: "Second signal.",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			session := NewSession(test.initial...)

			if test.build != nil {
				test.build(session)
			}

			got := session.Messages()
			if !slices.Equal(got, test.want) {
				t.Fatalf("Messages() = %#v, want %#v", got, test.want)
			}

			if len(got) > 0 {
				got[0] = Message{
					Role:    RoleAssistant,
					Content: "mutated returned history",
				}
			}

			if gotAgain := session.Messages(); !slices.Equal(gotAgain, test.want) {
				t.Errorf("Messages() after caller mutation = %#v, want %#v", gotAgain, test.want)
			}
		})
	}
}

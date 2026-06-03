package form

import "testing"

func TestValidate(t *testing.T) {
	valid := AdminForm{Title: "テスト", PublicSlug: "test-form", ProductIDs: []string{"1", "2"}}
	tests := []struct {
		name      string
		form      AdminForm
		wantError bool
	}{
		{name: "valid", form: valid},
		{name: "blank title", form: AdminForm{PublicSlug: "test"}, wantError: true},
		{name: "invalid slug", form: AdminForm{Title: "テスト", PublicSlug: "Test_Form"}, wantError: true},
		{name: "six products", form: AdminForm{Title: "テスト", PublicSlug: "test", ProductIDs: []string{"1", "2", "3", "4", "5", "6"}}, wantError: true},
		{name: "duplicate products", form: AdminForm{Title: "テスト", PublicSlug: "test", ProductIDs: []string{"1", "1"}}, wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := Validate(test.form, 5)
			if (err != nil) != test.wantError {
				t.Fatalf("Validate() error = %v, wantError = %v", err, test.wantError)
			}
		})
	}
}

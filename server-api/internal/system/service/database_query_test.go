package service

import "testing"

func TestValidateReadOnlyQuery(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{name: "simple select", query: "SELECT * FROM sys_users LIMIT 10"},
		{name: "comments and trailing semicolon", query: "/* read only */ SELECT id FROM sys_users LIMIT 5;"},
		{name: "keywords inside values", query: "SELECT 'delete update drop' AS content LIMIT 1"},
		{name: "quoted identifier", query: "SELECT `update` FROM demo LIMIT 1"},
		{name: "empty", query: "", wantErr: true},
		{name: "not select", query: "UPDATE sys_users SET username = 'x' LIMIT 1", wantErr: true},
		{name: "missing limit", query: "SELECT * FROM sys_users", wantErr: true},
		{name: "nested limit only", query: "SELECT * FROM (SELECT * FROM sys_users LIMIT 1) AS users", wantErr: true},
		{name: "multiple statements", query: "SELECT 1 LIMIT 1; DELETE FROM sys_users", wantErr: true},
		{name: "select into outfile", query: "SELECT * INTO OUTFILE '/tmp/users' FROM sys_users LIMIT 1", wantErr: true},
		{name: "select for update", query: "SELECT * FROM sys_users LIMIT 1 FOR UPDATE", wantErr: true},
		{name: "sleep function", query: "SELECT SLEEP(1) LIMIT 1", wantErr: true},
		{name: "assignment", query: "SELECT @value := 1 LIMIT 1", wantErr: true},
		{name: "executable comment", query: "SELECT 1 /*!50000 INTO OUTFILE '/tmp/x' */ LIMIT 1", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateReadOnlyQuery(test.query)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateReadOnlyQuery() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

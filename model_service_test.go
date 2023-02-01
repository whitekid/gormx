package gormx

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type MyModel struct {
	gorm.Model
}

func TestListModel(t *testing.T) {
	forEachSQLDriver(t, func(t *testing.T, dbURL string, reset func()) {
		db, err := Open(dbURL)
		require.NoError(t, err)

		db.AutoMigrate(&MyModel{})
		db.Save(&MyModel{})

		type args struct {
			opt ListOpt
		}
		tests := [...]struct {
			name    string
			args    args
			wantErr bool
		}{
			{`full list`, args{ListOpt{}}, false},
			{`count list`, args{ListOpt{Count: 5}}, false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				s := NewModelService[MyModel](db)
				got, err := s.List(nil, nil, tt.args.opt)
				require.Truef(t, (err != nil) == tt.wantErr, `List() failed: error = %+v, wantErr = %v`, err, tt.wantErr)
				if tt.wantErr {
					return
				}
				require.NotEmpty(t, got)
			})
		}

	})
}

package http

import "testing"

func Test_route_getOAPI(t *testing.T) {
	type fields struct {
		path   string
		method string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "simple path",
			fields: fields{
				path: "/users",
			},
			want: "/users",
		},
		{
			name: "simple path with id",
			fields: fields{
				path: "/users/:id",
			},
			want: "/users/{id}",
		},
		{
			name: "path with nested id",
			fields: fields{
				path: "/users/:id/password",
			},
			want: "/users/{id}/password",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := route{
				path:   tt.fields.path,
				method: tt.fields.method,
			}
			if got := r.getOAPI(); got != tt.want {
				t.Errorf("route.getOAPI() = %v, want %v", got, tt.want)
			}
		})
	}
}

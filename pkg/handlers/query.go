package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"reflect"

	gabi "github.com/app-sre/gabi/pkg"
)

type QueryData struct {
	Query string
}

func Query(env *gabi.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for k, v := range r.Header {
			env.Logger.Infof("Header field %q, Value %q", k, v)
		}

		var q QueryData
		err := json.NewDecoder(r.Body).Decode(&q)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		env.Logger.Infof("Query %q", q.Query)

		rows, err := env.DB.Query(q.Query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		defer rows.Close()
		cols, err := rows.Columns() // Remember to check err afterwards
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		vals := make([]interface{}, len(cols))
		var result [][]string
		var keys []string
		for i := range cols {
			vals[i] = new(sql.RawBytes)
			keys = append(keys, cols[i])
		}
		result = append(result, keys)
		for rows.Next() {
			err = rows.Scan(vals...)
			// Now you can check each element of vals for nil-ness,
			// and you can use type introspection and type assertions
			// to fetch the column into a typed variable.
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			var row []string
			for _, value := range vals {
				content := reflect.ValueOf(value).Interface().(*sql.RawBytes)
				row = append(row, string(*content))
			}
			result = append(result, row)
		}
		err = rows.Err()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

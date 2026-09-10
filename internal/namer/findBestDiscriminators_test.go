package namer

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindBestDiscriminators(t *testing.T) {
	tests := []struct {
		name     string
		types    [][]string
		expected []string
		only     bool
	}{
		{
			name: "example_01",
			types: [][]string{
				{"a", "b"},
				{"b", "c"},
			},
			expected: []string{"a", "c"},
		},
		{
			name: "example_02",
			types: [][]string{
				{"a", "b"},
				{"a", "b"},
			},
			expected: []string{"", ""},
		},
		{
			name: "example_03",
			types: [][]string{
				{"a", "b", "c", "d", "e"},
			},
			expected: []string{""},
		},
		{
			name: "example_04",
			types: [][]string{
				{"a", "b"},
				{"b", "c"},
				{"c", "d"},
			},
			expected: []string{"a", "b_c", "d"},
		},
		{
			name: "example_05",
			types: [][]string{
				{"a", "b", "c", "d"},
				{"b", "c", "d", "e"},
				{"c", "d", "e", "f"},
			},
			expected: []string{"a", "b_e", "f"},
		},
		{
			name: "example_06",
			types: [][]string{
				{"a"},
				{"b", "c"},
				{"b", "c"},
			},
			expected: []string{"a", "b", "b"},
		},
		{
			name: "example_07",
			types: [][]string{
				{"a", "d", "e", "f", "g"},
				{"b", "c"},
				{"b", "c"},
			},
			expected: []string{"a", "b", "b"},
		},
		{
			name: "example_08",
			types: [][]string{
				{"a", "b", "c"},
				{"d", "e", "f"},
				{"g", "h", "i"},
			},
			expected: []string{"a", "d", "g"},
		},
		{
			name: "example_09",
			types: [][]string{
				{"a", "b"},
				{"a", "b"},
				{"c", "d"},
				{"c", "d"},
			},
			expected: []string{"a", "a", "c", "c"},
		},
		{
			name: "example_10",
			types: [][]string{
				{"a", "b"},
				{"b", "c"},
				{"d", "e"},
				{"e", "f"},
			},
			expected: []string{"a", "c", "d", "f"},
		},
		{
			name: "example_11",
			types: [][]string{
				{"a", "b", "c", "d"},
				{"e", "b", "c", "d"},
				{"f", "g", "h", "i"},
				{"j", "g", "h", "i"},
			},
			expected: []string{"a", "e", "f", "j"},
		},
		{
			name: "example_12",
			types: [][]string{
				{"b", "a", "c", "g"},
				{"c", "e", "b", "h"},
			},
			expected: []string{"a", "e"},
		},
		{
			name: "example_13",
			types: [][]string{
				{"a", "b", "c"},
				{"c", "b", "a"},
			},
			expected: []string{"", ""},
		},
		{
			name: "example_14",
			types: [][]string{
				{"a", "b"},
				{"b"},
				{"c", "d"},
			},
			expected: []string{"a", "b", "c"},
		},
		{
			name: "example_14_a",
			types: [][]string{
				{"a", "b", "c", "d", "e", "f", "g", "h", "i"},
				{"a", "b", "c", "d", "e", "f", "g", "h"},
				{"A"},
			},
			expected: []string{"i", "a", "A"},
		},
		{
			name: "example_15",
			types: [][]string{
				{"a", "b"},
				{"b"},
			},
			expected: []string{"a", ""},
		},
		{
			name: "example_15_a",
			types: [][]string{
				{"b", "a"},
				{"b"},
			},
			expected: []string{"a", ""},
		},
		{
			name: "example_15_b",
			types: [][]string{
				{"shared", "Schemas"},
				{"shared", "Schemas", "OneOfWithFactoredOutProperties"},
			},
			expected: []string{"", "OneOfWithFactoredOutProperties"},
		},
		{
			name: "example_16",
			types: [][]string{
				{"a"},
				{"a"},
				{"a"},
				{"a"},
				{"b"},
				{"c"},
				{"d"},
				{"e"},
			},
			expected: []string{"a", "a", "a", "a", "b", "c", "d", "e"},
		},
		{
			name: "example_17",
			types: [][]string{
				{"errors", "enum", "createOrder", "BadRequest", "Bad", "request", "responseBody", "ApplicationJSON", "400", "response"},
				{"errors", "enum", "createOrder", "Forbidden", "responseBody", "ApplicationJSON", "403", "response"},
				{"errors", "enum", "createOrder", "Unauthorized", "responseBody", "ApplicationJSON", "401", "response"},
				{"errors", "enum", "createReview", "BadRequest", "Bad", "request", "responseBody", "ApplicationJSON", "400", "response"},
				{"errors", "enum", "createReview", "Forbidden", "responseBody", "ApplicationJSON", "403", "response"},
				{"errors", "enum", "createReview", "Unauthorized", "responseBody", "ApplicationJSON", "401", "response"},
				{"errors", "enum", "createUser", "BadRequest", "Bad", "request", "responseBody", "ApplicationJSON", "400", "response"},
				{"errors", "enum", "createUser", "Forbidden", "responseBody", "ApplicationJSON", "403", "response"},
				{"errors", "enum", "createUser", "Unauthorized", "responseBody", "ApplicationJSON", "401", "response"},
				{"errors", "enum", "getCart", "BadRequest", "Bad", "request", "responseBody", "ApplicationJSON", "400", "response"},
				{"errors", "enum", "getCart", "Forbidden", "responseBody", "ApplicationJSON", "403", "response"},
				{"errors", "enum", "getCart", "Unauthorized", "responseBody", "ApplicationJSON", "401", "response"},
				{"errors", "enum", "getNotifications", "BadRequest", "Bad", "request", "responseBody", "ApplicationJSON", "400", "response"},
				{"errors", "enum", "getNotifications", "Forbidden", "responseBody", "ApplicationJSON", "403", "response"},
				{"errors", "enum", "getNotifications", "Unauthorized", "responseBody", "ApplicationJSON", "401", "response"},
				{"errors", "enum", "getUser", "BadRequest", "Bad", "request", "responseBody", "ApplicationJSON", "400", "response"},
				{"errors", "enum", "getUser", "Forbidden", "responseBody", "ApplicationJSON", "403", "response"},
				{"errors", "enum", "getUser", "Unauthorized", "responseBody", "ApplicationJSON", "401", "response"},
				{"errors", "enum", "getWishlist", "BadRequest", "Bad", "request", "responseBody", "ApplicationJSON", "400", "response"},
				{"errors", "enum", "getWishlist", "Forbidden", "responseBody", "ApplicationJSON", "403", "response"},
				{"errors", "enum", "getWishlist", "Unauthorized", "responseBody", "ApplicationJSON", "401", "response"},
				{"errors", "enum", "listCategories", "BadRequest", "Bad", "request", "responseBody", "ApplicationJSON", "400", "response"},
				{"errors", "enum", "listCategories", "Forbidden", "responseBody", "ApplicationJSON", "403", "response"},
				{"errors", "enum", "listCategories", "Unauthorized", "responseBody", "ApplicationJSON", "401", "response"},
				{"errors", "enum", "listProducts", "BadRequest", "Bad", "request", "responseBody", "ApplicationJSON", "400", "response"},
				{"errors", "enum", "listProducts", "Forbidden", "responseBody", "ApplicationJSON", "403", "response"},
				{"errors", "enum", "listProducts", "Unauthorized", "responseBody", "ApplicationJSON", "401", "response"},
				{"errors", "enum", "listUsers", "BadRequest", "Bad", "request", "responseBody", "ApplicationJSON", "400", "response"},
				{"errors", "enum", "listUsers", "Forbidden", "responseBody", "ApplicationJSON", "403", "response"},
				{"errors", "enum", "listUsers", "Unauthorized", "responseBody", "ApplicationJSON", "401", "response"},
			},
			expected: []string{
				"BadRequest_createOrder",
				"Forbidden_createOrder",
				"Unauthorized_createOrder",
				"BadRequest_createReview",
				"Forbidden_createReview",
				"Unauthorized_createReview",
				"BadRequest_createUser",
				"Forbidden_createUser",
				"Unauthorized_createUser",
				"BadRequest_getCart",
				"Forbidden_getCart",
				"Unauthorized_getCart",
				"BadRequest_getNotifications",
				"Forbidden_getNotifications",
				"Unauthorized_getNotifications",
				"BadRequest_getUser",
				"Forbidden_getUser",
				"Unauthorized_getUser",
				"BadRequest_getWishlist",
				"Forbidden_getWishlist",
				"Unauthorized_getWishlist",
				"BadRequest_listCategories",
				"Forbidden_listCategories",
				"Unauthorized_listCategories",
				"BadRequest_listProducts",
				"Forbidden_listProducts",
				"Unauthorized_listProducts",
				"BadRequest_listUsers",
				"Forbidden_listUsers",
				"Unauthorized_listUsers",
			},
		},
		{
			name: "example_18",
			types: [][]string{
				{"a", "b", "c"},
				{"A", "a", "c"},
				{"B", "a", "b"},
				{"C", "b", "c"},
			},
			expected: []string{"a_b_c", "A", "B", "C"},
		},
		{

			name: "example_18_a",
			types: [][]string{
				{"a", "b", "c"},
				{"A", "B", "a", "c"},
				{"B", "C", "a", "b"},
				{"C", "D", "b", "c"},
			},
			expected: []string{"a_b_c", "A", "B_C", "D"},
		},
		{
			name: "example_19",
			types: [][]string{
				{"a", "b", "c"},
				{"a", "b", "d"},
				{"a", "c", "d"},
				{"b", "c", "d"},
			},
			expected: []string{
				"a_b_c",
				"a_b_d",
				"a_c_d",
				"b_c_d",
			},
		},
		{
			// This is for ensuring performance rather than correctness
			name: "example_20",
			types: [][]string{
				// a-z excluding a
				{"b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"},
				// a-z excluding b
				{"a", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"},
				// a-z excluding c
				{"a", "b", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"},
				// a-z excluding d
				{"a", "b", "c", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"},
				// a-z excluding e
				{"a", "b", "c", "d", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"},
				// a-z excluding f
				{"a", "b", "c", "d", "e", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"},
				// a-z excluding g
				{"a", "b", "c", "d", "e", "f", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"},
				// a-z excluding h
				{"a", "b", "c", "d", "e", "f", "g", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"},
				// a-z excluding i
				{"a", "b", "c", "d", "e", "f", "g", "h", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"},
				// a-z excluding j
				{"a", "b", "c", "d", "e", "f", "g", "h", "i", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"},
				// a-z excluding k
				{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"},
				// a-z excluding l
				{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"},
				// a-z excluding m
				{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"},
				// a-z excluding n
				{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"},
				// a-z excluding o
				{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"},
				// a-z excluding p
				{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"},
				// a-z excluding q
				{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "r", "s", "t", "u", "v", "w", "x", "y", "z"},
				// a-z excluding r
				{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "s", "t", "u", "v", "w", "x", "y", "z"},
				// a-z excluding s
				{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "t", "u", "v", "w", "x", "y", "z"},
				// a-z excluding t
				{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "u", "v", "w", "x", "y", "z"},
				// a-z excluding u
				{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "v", "w", "x", "y", "z"},
				// a-z excluding v
				{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "w", "x", "y", "z"},
				// a-z excluding w
				{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "x", "y", "z"},
				// a-z excluding x
				{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "y", "z"},
				// a-z excluding y
				{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "z"},
				// a-z excluding z
				{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y"},
			},
			expected: []string{
				"b_c_d_e_f_g_h_i_j_k_l_m_n_o_p_q_r_s_t_u_v_w_x_y_z",
				"a_c_d_e_f_g_h_i_j_k_l_m_n_o_p_q_r_s_t_u_v_w_x_y_z",
				"a_b_d_e_f_g_h_i_j_k_l_m_n_o_p_q_r_s_t_u_v_w_x_y_z",
				"a_b_c_e_f_g_h_i_j_k_l_m_n_o_p_q_r_s_t_u_v_w_x_y_z",
				"a_b_c_d_f_g_h_i_j_k_l_m_n_o_p_q_r_s_t_u_v_w_x_y_z",
				"a_b_c_d_e_g_h_i_j_k_l_m_n_o_p_q_r_s_t_u_v_w_x_y_z",
				"a_b_c_d_e_f_h_i_j_k_l_m_n_o_p_q_r_s_t_u_v_w_x_y_z",
				"a_b_c_d_e_f_g_i_j_k_l_m_n_o_p_q_r_s_t_u_v_w_x_y_z",
				"a_b_c_d_e_f_g_h_j_k_l_m_n_o_p_q_r_s_t_u_v_w_x_y_z",
				"a_b_c_d_e_f_g_h_i_k_l_m_n_o_p_q_r_s_t_u_v_w_x_y_z",
				"a_b_c_d_e_f_g_h_i_j_l_m_n_o_p_q_r_s_t_u_v_w_x_y_z",
				"a_b_c_d_e_f_g_h_i_j_k_m_n_o_p_q_r_s_t_u_v_w_x_y_z",
				"a_b_c_d_e_f_g_h_i_j_k_l_n_o_p_q_r_s_t_u_v_w_x_y_z",
				"a_b_c_d_e_f_g_h_i_j_k_l_m_o_p_q_r_s_t_u_v_w_x_y_z",
				"a_b_c_d_e_f_g_h_i_j_k_l_m_n_p_q_r_s_t_u_v_w_x_y_z",
				"a_b_c_d_e_f_g_h_i_j_k_l_m_n_o_q_r_s_t_u_v_w_x_y_z",
				"a_b_c_d_e_f_g_h_i_j_k_l_m_n_o_p_r_s_t_u_v_w_x_y_z",
				"a_b_c_d_e_f_g_h_i_j_k_l_m_n_o_p_q_s_t_u_v_w_x_y_z",
				"a_b_c_d_e_f_g_h_i_j_k_l_m_n_o_p_q_r_t_u_v_w_x_y_z",
				"a_b_c_d_e_f_g_h_i_j_k_l_m_n_o_p_q_r_s_u_v_w_x_y_z",
				"a_b_c_d_e_f_g_h_i_j_k_l_m_n_o_p_q_r_s_t_v_w_x_y_z",
				"a_b_c_d_e_f_g_h_i_j_k_l_m_n_o_p_q_r_s_t_u_w_x_y_z",
				"a_b_c_d_e_f_g_h_i_j_k_l_m_n_o_p_q_r_s_t_u_v_x_y_z",
				"a_b_c_d_e_f_g_h_i_j_k_l_m_n_o_p_q_r_s_t_u_v_w_y_z",
				"a_b_c_d_e_f_g_h_i_j_k_l_m_n_o_p_q_r_s_t_u_v_w_x_z",
				"a_b_c_d_e_f_g_h_i_j_k_l_m_n_o_p_q_r_s_t_u_v_w_x_y",
			},
		},
	}

	hasOnly := false
	for _, tt := range tests {
		if tt.only {
			hasOnly = true
			break
		}
	}

	for _, tt := range tests {
		if hasOnly && !tt.only {
			continue
		}

		t.Run(tt.name, func(t *testing.T) {

			fmt.Println(">>>>> Example", tt.name, tt.types, tt.expected)
			for _, t := range tt.types {
				fmt.Println("Type", t, "->", strings.Join(t, "_"))
			}
			res := FindBestDiscriminators(tt.types)

			got := []string{}
			for _, e := range res {
				got = append(got, strings.Join(e, "_"))
			}

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("For test %q, want %v, got %v", tt.name, tt.expected, got)
			}

		})
	}
}

func TestTieBreaker(t *testing.T) {
	tests := []struct {
		name       string
		types      [][]string
		tieBreaker TieBreaker
		expected   []string
	}{
		{
			name: "example_01",
			types: [][]string{
				{"x"},
				{"y", "z"},
			},
			tieBreaker: func(a, b string) string {
				if a == "y" {
					return a
				}
				return b
			},
			expected: []string{"x", "y"},
		},
		{
			name: "example_02",
			types: [][]string{
				{"x"},
				{"y", "z"},
			},
			tieBreaker: func(a, b string) string {
				if a == "z" {
					return a
				}
				return b
			},
			expected: []string{"x", "z"},
		},
		{
			name: "example_03",
			types: [][]string{
				{"a", "b", "c", "d", "e"},
				{"f", "g", "h", "i", "j"},
			},
			tieBreaker: func(a, b string) string {
				return ""
			},
			expected: []string{"a", "f"},
		},
		{
			name: "example_04",
			types: [][]string{
				{"a", "b", "c", "d", "e"},
				{"f", "g", "h", "i", "j"},
			},
			tieBreaker: func(a, b string) string {
				// Prefer higher alphabetical order
				if a > b {
					return a
				}
				return b
			},
			expected: []string{"e", "j"},
		},
		{
			name: "example_05",
			types: [][]string{
				{"a", "b", "c", "d", "e"},
				{"f", "g", "h", "i", "j"},
			},
			tieBreaker: func(a, b string) string {
				// Prefer lower alphabetical order
				if a < b {
					return a
				}
				return b
			},
			expected: []string{"a", "f"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := FindBestDiscriminators(tt.types, tt.tieBreaker)

			expected := [][]string{}
			for _, e := range tt.expected {
				expected = append(expected, strings.Split(e, "_"))
			}

			assert.Equal(t, expected, res)
		})
	}
}

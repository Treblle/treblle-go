package treblle

import "reflect"

func (s *TestSuite) Test_Masking() {
	testCases := map[string]struct {
		input          []byte
		expectedErr    error
		expectedOutput map[string]interface{}
	}{
		"happy-path": {
			input:          []byte(`{"id":2}`),
			expectedErr:    nil,
			expectedOutput: map[string]interface{}{"id": 2.0},
		},
		"invalid-json": {
			input:          []byte(`{"id":`),
			expectedErr:    ErrNotJson,
			expectedOutput: nil,
		},
		"mask-password": {
			input:          []byte(`{"id":2,"password":"test123"}`),
			expectedErr:    nil,
			expectedOutput: map[string]interface{}{"id": 2.0, "password": "*******"},
		},
		"mask-any-level": {
			input:          []byte(`{"id":2,"node":{"password":"test123"}}`),
			expectedErr:    nil,
			expectedOutput: map[string]interface{}{"id": 2.0, "node": map[string]interface{}{"password": "*******"}},
		},
	}

	for tn, tc := range testCases {
		Configure(Configuration{FieldsToMask: []string{"password"}})
		masked, err := getMaskedJSON(tc.input)
		if tc.expectedErr != nil {
			s.Require().Error(err, tn)
			s.Require().Equal(tc.expectedErr, err, tn)
		} else {
			s.Require().True(reflect.DeepEqual(tc.expectedOutput, masked), tn)
		}
	}
}

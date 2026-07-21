package util

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

// TestMapAlternateKeysValue a tiny test for util.MapAlternateKeysValue()
func TestMapAlternateKeysValue(t *testing.T) {
	keyVal := "XXXXX"
	altKey := "aws_access_key_id"
	canonicalKey := "AWS_ACCESS_KEY_ID"

	actualAltValue := MapAlternateKeysValue(map[string]string{
		altKey: keyVal,
	}, canonicalKey)
	if actualAltValue != keyVal {
		t.Fatalf("Alternative key: expected %q, got %q", keyVal, actualAltValue)
	}

	actualCanonicalValue := MapAlternateKeysValue(map[string]string{
		canonicalKey: keyVal,
	}, canonicalKey)
	if actualCanonicalValue != keyVal {
		t.Fatalf("Canonical key: expected %q, got %q", keyVal, actualCanonicalValue)
	}
}

// TestIsTolerationSuperset tests the IsTolerationSuperset function
func TestIsTolerationSuperset(t *testing.T) {
	sec60 := int64(60)
	sec30 := int64(30)

	equal := corev1.Toleration{
		Key:      "k",
		Operator: corev1.TolerationOpEqual,
		Value:    "v",
		Effect:   corev1.TaintEffectNoSchedule,
	}
	emptyOp := corev1.Toleration{
		Key:    "k",
		Value:  "v",
		Effect: corev1.TaintEffectNoSchedule,
	}
	exists := corev1.Toleration{
		Operator: corev1.TolerationOpExists,
	}
	keyExists := corev1.Toleration{
		Key:      "k",
		Operator: corev1.TolerationOpExists,
		Effect:   corev1.TaintEffectNoSchedule,
	}
	other := corev1.Toleration{
		Key:      "other",
		Operator: corev1.TolerationOpEqual,
		Value:    "v",
		Effect:   corev1.TaintEffectNoSchedule,
	}
	noExecute60 := corev1.Toleration{
		Key:               "k",
		Operator:          corev1.TolerationOpExists,
		Effect:            corev1.TaintEffectNoExecute,
		TolerationSeconds: &sec60,
	}
	noExecute30 := corev1.Toleration{
		Key:               "k",
		Operator:          corev1.TolerationOpExists,
		Effect:            corev1.TaintEffectNoExecute,
		TolerationSeconds: &sec30,
	}
	noExecuteForever := corev1.Toleration{
		Key:      "k",
		Operator: corev1.TolerationOpExists,
		Effect:   corev1.TaintEffectNoExecute,
	}

	tests := []struct {
		name     string
		ss       corev1.Toleration
		target   corev1.Toleration
		expected bool
	}{
		{
			name:     "identical",
			ss:       equal,
			target:   equal,
			expected: true,
		},
		{
			name:     "bare Exists covers Equal",
			ss:       exists,
			target:   equal,
			expected: true,
		},
		{
			name:     "Equal does not cover bare Exists",
			ss:       equal,
			target:   exists,
			expected: false,
		},
		{
			name:     "key Exists covers Equal",
			ss:       keyExists,
			target:   equal,
			expected: true,
		},
		{
			name:     "unrelated key",
			ss:       other,
			target:   equal,
			expected: false,
		},
		{
			name:     "empty ss operator means Equal",
			ss:       emptyOp,
			target:   equal,
			expected: true,
		},
		{
			name:     "empty target operator means Equal",
			ss:       equal,
			target:   emptyOp,
			expected: true,
		},
		{
			name:     "both empty operators match",
			ss:       emptyOp,
			target:   emptyOp,
			expected: true,
		},
		{
			name: "effect mismatch",
			ss: corev1.Toleration{
				Key:      "k",
				Operator: corev1.TolerationOpEqual,
				Value:    "v",
				Effect:   corev1.TaintEffectNoSchedule,
			},
			target: corev1.Toleration{
				Key:      "k",
				Operator: corev1.TolerationOpEqual,
				Value:    "v",
				Effect:   corev1.TaintEffectNoExecute,
			},
			expected: false,
		},
		{
			name:     "NoExecute with shorter seconds does not cover forever",
			ss:       noExecute60,
			target:   noExecuteForever,
			expected: false,
		},
		{
			name:     "NoExecute with shorter seconds does not cover longer seconds",
			ss:       noExecute30,
			target:   noExecute60,
			expected: false,
		},
		{
			name:     "NoExecute covers equal or shorter seconds",
			ss:       noExecute60,
			target:   noExecute30,
			expected: true,
		},
		{
			name: "unknown operator",
			ss: corev1.Toleration{
				Key:      "k",
				Operator: "Unknown",
				Effect:   corev1.TaintEffectNoSchedule,
			},
			target: corev1.Toleration{
				Key:      "k",
				Operator: corev1.TolerationOpEqual,
				Value:    "v",
				Effect:   corev1.TaintEffectNoSchedule,
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := IsTolerationSuperset(tc.ss, tc.target)
			if actual != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, actual)
			}
		})
	}
}

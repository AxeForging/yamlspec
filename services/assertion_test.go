package services

import (
	"testing"

	"github.com/AxeForging/yamlspec/domain"
)

func ptr[T any](v T) *T { return &v }

func deploymentManifest() []interface{} {
	manifests, _ := ParseManifests(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  namespace: production
  labels:
    app: my-app
    tier: frontend
    app.kubernetes.io/name: my-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app
  template:
    spec:
      containers:
        - name: app
          image: "nginx:1.25.3"
          ports:
            - containerPort: 80
            - containerPort: 443
          resources:
            limits:
              cpu: "500m"
              memory: "256Mi"
            requests:
              cpu: "100m"
              memory: "128Mi"
          env:
            - name: NODE_ENV
              value: production
            - name: LOG_LEVEL
              value: info
---
apiVersion: v1
kind: Service
metadata:
  name: my-app-svc
  namespace: production
spec:
  type: ClusterIP
  ports:
    - port: 80
      targetPort: 80
  selector:
    app: my-app
`)
	return manifests
}

func TestEvaluate_Selector(t *testing.T) {
	ae := NewAssertionEngine()
	manifests := deploymentManifest()

	t.Run("selects Deployment", func(t *testing.T) {
		describes := []domain.DescBlock{{
			Name:   "Deployment",
			Select: `select(.kind == "Deployment")`,
			It: []domain.Assertion{{
				Should:  "have correct name",
				Expect:  "metadata.name",
				ToEqual: "my-app",
			}},
		}}
		results := ae.Evaluate(describes, manifests)
		assertStatus(t, results[0].Status, domain.StatusPassed)
	})

	t.Run("selects Service", func(t *testing.T) {
		describes := []domain.DescBlock{{
			Name:   "Service",
			Select: `select(.kind == "Service")`,
			It: []domain.Assertion{{
				Should:  "have correct name",
				Expect:  "metadata.name",
				ToEqual: "my-app-svc",
			}},
		}}
		results := ae.Evaluate(describes, manifests)
		assertStatus(t, results[0].Status, domain.StatusPassed)
	})

	t.Run("empty selector returns all", func(t *testing.T) {
		describes := []domain.DescBlock{{
			Name: "All",
			It: []domain.Assertion{{
				Should:  "have metadata",
				Expect:  "metadata.name",
				ToExist: ptr(true),
			}},
		}}
		results := ae.Evaluate(describes, manifests)
		assertStatus(t, results[0].Status, domain.StatusPassed)
	})

	t.Run("no match fails", func(t *testing.T) {
		describes := []domain.DescBlock{{
			Name:   "Missing",
			Select: `select(.kind == "StatefulSet")`,
			It:     []domain.Assertion{{Should: "exist", Expect: "metadata.name", ToExist: ptr(true)}},
		}}
		results := ae.Evaluate(describes, manifests)
		assertStatus(t, results[0].Status, domain.StatusFailed)
	})
}

func TestEvaluate_ToEqual(t *testing.T) {
	ae := NewAssertionEngine()
	manifests := deploymentManifest()

	t.Run("string equal passes", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "equal", Expect: "metadata.name", ToEqual: "my-app",
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("string equal fails", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "equal", Expect: "metadata.name", ToEqual: "wrong",
		})
		assertStatus(t, r.Status, domain.StatusFailed)
	})

	t.Run("int equal passes", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "equal", Expect: "spec.replicas", ToEqual: 3,
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})
}

func TestEvaluate_ToNotEqual(t *testing.T) {
	ae := NewAssertionEngine()
	manifests := deploymentManifest()

	t.Run("passes when different", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "not equal", Expect: "metadata.name", ToNotEqual: "wrong",
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("fails when same", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "not equal", Expect: "metadata.name", ToNotEqual: "my-app",
		})
		assertStatus(t, r.Status, domain.StatusFailed)
	})
}

func TestEvaluate_NumericComparisons(t *testing.T) {
	ae := NewAssertionEngine()
	manifests := deploymentManifest()

	t.Run("toBeGreaterThan passes", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "gt", Expect: "spec.replicas", ToBeGreaterThan: ptr(2.0),
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("toBeGreaterThan fails", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "gt", Expect: "spec.replicas", ToBeGreaterThan: ptr(3.0),
		})
		assertStatus(t, r.Status, domain.StatusFailed)
	})

	t.Run("toBeLessThan passes", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "lt", Expect: "spec.replicas", ToBeLessThan: ptr(5.0),
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("toBeLessThan fails", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "lt", Expect: "spec.replicas", ToBeLessThan: ptr(3.0),
		})
		assertStatus(t, r.Status, domain.StatusFailed)
	})

	t.Run("toBeGreaterOrEqual passes on equal", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "gte", Expect: "spec.replicas", ToBeGreaterOrEqual: ptr(3.0),
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("toBeLessOrEqual passes on equal", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "lte", Expect: "spec.replicas", ToBeLessOrEqual: ptr(3.0),
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("toBeGreaterOrEqual fails", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "gte", Expect: "spec.replicas", ToBeGreaterOrEqual: ptr(5.0),
		})
		assertStatus(t, r.Status, domain.StatusFailed)
	})

	t.Run("toBeLessOrEqual fails", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "lte", Expect: "spec.replicas", ToBeLessOrEqual: ptr(2.0),
		})
		assertStatus(t, r.Status, domain.StatusFailed)
	})
}

func TestEvaluate_StringOperators(t *testing.T) {
	ae := NewAssertionEngine()
	manifests := deploymentManifest()

	t.Run("toContain passes", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "contain", Expect: "spec.template.spec.containers[0].image", ToContain: "nginx",
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("toContain fails", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "contain", Expect: "spec.template.spec.containers[0].image", ToContain: "alpine",
		})
		assertStatus(t, r.Status, domain.StatusFailed)
	})

	t.Run("toNotContain passes", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "not contain", Expect: "spec.template.spec.containers[0].image", ToNotContain: "latest",
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("toStartWith passes", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "start", Expect: "spec.template.spec.containers[0].image", ToStartWith: "nginx:",
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("toEndWith passes", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "end", Expect: "spec.template.spec.containers[0].image", ToEndWith: "1.25.3",
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("toNotStartWith passes", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "not start", Expect: "spec.template.spec.containers[0].image", ToNotStartWith: "alpine:",
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("toNotEndWith passes", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "not end", Expect: "spec.template.spec.containers[0].image", ToNotEndWith: ":latest",
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("toMatch passes", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "match", Expect: "spec.template.spec.containers[0].image", ToMatch: `^nginx:\d+\.\d+\.\d+$`,
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("toMatch fails", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "match", Expect: "spec.template.spec.containers[0].image", ToMatch: `^alpine:`,
		})
		assertStatus(t, r.Status, domain.StatusFailed)
	})
}

func TestEvaluate_Existence(t *testing.T) {
	ae := NewAssertionEngine()
	manifests := deploymentManifest()

	t.Run("toExist true passes for existing field", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "exist", Expect: "spec.replicas", ToExist: ptr(true),
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("toExist true fails for missing field", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "exist", Expect: "spec.nonexistent", ToExist: ptr(true),
		})
		assertStatus(t, r.Status, domain.StatusFailed)
	})

	t.Run("toExist false passes for missing field", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "not exist", Expect: "spec.nonexistent", ToExist: ptr(false),
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("toExist false fails for existing field", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "not exist", Expect: "spec.replicas", ToExist: ptr(false),
		})
		assertStatus(t, r.Status, domain.StatusFailed)
	})
}

func TestEvaluate_ToHaveKey(t *testing.T) {
	ae := NewAssertionEngine()
	manifests := deploymentManifest()

	t.Run("passes for existing key", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "have key", Expect: "metadata.labels", ToHaveKey: "app",
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("fails for missing key", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "have key", Expect: "metadata.labels", ToHaveKey: "missing",
		})
		assertStatus(t, r.Status, domain.StatusFailed)
	})
}

func TestEvaluate_SetMembership(t *testing.T) {
	ae := NewAssertionEngine()
	manifests := deploymentManifest()

	t.Run("toBeOneOf passes", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "one of", Expect: "metadata.namespace",
			ToBeOneOf: []interface{}{"production", "staging"},
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("toBeOneOf fails", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "one of", Expect: "metadata.namespace",
			ToBeOneOf: []interface{}{"staging", "dev"},
		})
		assertStatus(t, r.Status, domain.StatusFailed)
	})

	t.Run("toNotBeOneOf passes", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "not one of", Expect: "metadata.namespace",
			ToNotBeOneOf: []interface{}{"staging", "dev"},
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("toNotBeOneOf fails", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "not one of", Expect: "metadata.namespace",
			ToNotBeOneOf: []interface{}{"production", "staging"},
		})
		assertStatus(t, r.Status, domain.StatusFailed)
	})
}

func TestEvaluate_ArrayOperators(t *testing.T) {
	ae := NewAssertionEngine()
	manifests := deploymentManifest()

	t.Run("toHaveLength passes", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "length", Expect: "spec.template.spec.containers[0].ports", ToHaveLength: ptr(2),
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("toHaveLength fails", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "length", Expect: "spec.template.spec.containers[0].ports", ToHaveLength: ptr(3),
		})
		assertStatus(t, r.Status, domain.StatusFailed)
	})

	t.Run("toHaveMinLength passes", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "min length", Expect: "spec.template.spec.containers[0].ports", ToHaveMinLength: ptr(1),
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("toHaveMinLength fails", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "min length", Expect: "spec.template.spec.containers[0].ports", ToHaveMinLength: ptr(5),
		})
		assertStatus(t, r.Status, domain.StatusFailed)
	})

	t.Run("toHaveMaxLength passes", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "max length", Expect: "spec.template.spec.containers[0].ports", ToHaveMaxLength: ptr(5),
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("toHaveMaxLength fails", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "max length", Expect: "spec.template.spec.containers[0].ports", ToHaveMaxLength: ptr(1),
		})
		assertStatus(t, r.Status, domain.StatusFailed)
	})

	t.Run("toContainItem passes with int", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Service", domain.Assertion{
			Should: "contain", Expect: "spec.ports[0].port", ToEqual: 80,
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})
}

func TestEvaluate_FieldPaths(t *testing.T) {
	ae := NewAssertionEngine()
	manifests := deploymentManifest()

	t.Run("nested field path", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "access nested", Expect: "spec.template.spec.containers[0].resources.limits.cpu",
			ToEqual: "500m",
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("bracket notation for special keys", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should:  "bracket notation",
			Expect:  `.metadata.labels["app.kubernetes.io/name"]`,
			ToEqual: "my-app",
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("leading dot path", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "leading dot", Expect: ".spec.replicas", ToEqual: 3,
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("missing field returns not exist", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should: "not exist", Expect: "spec.nonexistent.deep.path", ToExist: ptr(false),
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})
}

func TestEvaluate_MultipleOperators(t *testing.T) {
	ae := NewAssertionEngine()
	manifests := deploymentManifest()

	t.Run("range check passes", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should:             "be in range",
			Expect:             "spec.replicas",
			ToBeGreaterOrEqual: ptr(1.0),
			ToBeLessOrEqual:    ptr(5.0),
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})

	t.Run("string with multiple checks", func(t *testing.T) {
		r := evalSingle(ae, manifests, "Deployment", domain.Assertion{
			Should:      "match constraints",
			Expect:      "spec.template.spec.containers[0].image",
			ToStartWith: "nginx:",
			ToNotEndWith: ":latest",
			ToContain:   "1.25",
		})
		assertStatus(t, r.Status, domain.StatusPassed)
	})
}

func TestParseManifests(t *testing.T) {
	t.Run("single document", func(t *testing.T) {
		m, err := ParseManifests("kind: Deployment\nmetadata:\n  name: test")
		if err != nil {
			t.Fatal(err)
		}
		if len(m) != 1 {
			t.Fatalf("expected 1 manifest, got %d", len(m))
		}
	})

	t.Run("multi document", func(t *testing.T) {
		m, err := ParseManifests("kind: Deployment\n---\nkind: Service\n---\nkind: ConfigMap")
		if err != nil {
			t.Fatal(err)
		}
		if len(m) != 3 {
			t.Fatalf("expected 3 manifests, got %d", len(m))
		}
	})

	t.Run("skips empty documents", func(t *testing.T) {
		m, err := ParseManifests("---\n---\nkind: Service\n---\n")
		if err != nil {
			t.Fatal(err)
		}
		if len(m) != 1 {
			t.Fatalf("expected 1 manifest, got %d", len(m))
		}
	})
}

// Helpers

func evalSingle(ae *AssertionEngine, manifests []interface{}, kind string, assertion domain.Assertion) domain.AssertionResult {
	selector := ""
	if kind != "" {
		selector = `select(.kind == "` + kind + `")`
	}
	describes := []domain.DescBlock{{
		Name:   kind,
		Select: selector,
		It:     []domain.Assertion{assertion},
	}}
	results := ae.Evaluate(describes, manifests)
	if len(results) == 0 || len(results[0].Assertions) == 0 {
		return domain.AssertionResult{Status: domain.StatusError, Error: "no results"}
	}
	return results[0].Assertions[0]
}

func assertStatus(t *testing.T, got, want domain.Status) {
	t.Helper()
	if got != want {
		t.Errorf("expected status %s, got %s", want, got)
	}
}

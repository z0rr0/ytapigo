package ytapigo

import (
    // "fmt"
    "testing"
)

func TestCheckYT(t *testing.T) {
    if TestMsg != CheckYT() {
    	t.Errorf("Failed simple test")
    }
}

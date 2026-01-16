package internal

import (
	"github.com/onsi/gomega"
	"testing"
)

func TestFormatQuotedList_Empty(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	list := []string{}
	g.Expect(FormatQuotedList(list)).To(gomega.Equal(""))
}

func TestFormatQuotedList_OneElement(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	list := []string{"a"}
	g.Expect(FormatQuotedList(list)).To(gomega.Equal(`"a"`))
}

func TestFormatQuotedList_TwoElements(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	list := []string{"a", "b"}
	g.Expect(FormatQuotedList(list)).To(gomega.Equal(`"a" or "b"`))
}

func TestFormatQuotedList_ThreeElements(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	list := []string{"a", "b", "c"}
	g.Expect(FormatQuotedList(list)).To(gomega.Equal(`"a", "b", or "c"`))
}

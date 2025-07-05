package instautoctrl_test

import (
	"context"
	"time"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/instautoctrl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Common", func() {
	var _ = Describe("ParseDurationWithDays", func() {
		var ctx context.Context

		BeforeEach(func() {
			ctx = context.TODO()
		})

		It("should correctly parse a valid day duration", func() {
			dur, err := instautoctrl.ParseDurationWithDays(ctx, "7d")
			Expect(err).NotTo(HaveOccurred())
			Expect(dur).To(Equal(7 * 24 * time.Hour))
		})

		It("should correctly parse a valid hour duration", func() {
			dur, err := instautoctrl.ParseDurationWithDays(ctx, "72h")
			Expect(err).NotTo(HaveOccurred())
			Expect(dur).To(Equal(72 * time.Hour))
		})

		It("should correctly parse a valid minute duration", func() {
			dur, err := instautoctrl.ParseDurationWithDays(ctx, "30m")
			Expect(err).NotTo(HaveOccurred())
			Expect(dur).To(Equal(30 * time.Minute))
		})

		It("should return an error for an invalid format", func() {
			_, err := instautoctrl.ParseDurationWithDays(ctx, "abc")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid input format"))
		})

		It("should return an error for missing unit", func() {
			_, err := instautoctrl.ParseDurationWithDays(ctx, "10")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid input format"))
		})

		It("should return an error for non-numeric day value", func() {
			_, err := instautoctrl.ParseDurationWithDays(ctx, "xd")
			Expect(err).To(HaveOccurred())
		})

		It("should return an error for invalid duration input", func() {
			_, err := instautoctrl.ParseDurationWithDays(ctx, "10x")
			Expect(err).To(HaveOccurred())
		})

		It("should return an error for 'never' if not supported", func() {
			_, err := instautoctrl.ParseDurationWithDays(ctx, "never")
			Expect(err).To(HaveOccurred())
		})
	})
})

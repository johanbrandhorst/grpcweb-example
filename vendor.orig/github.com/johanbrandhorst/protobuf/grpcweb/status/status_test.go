package status_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/codes"

	"github.com/johanbrandhorst/protobuf/grpcweb/status"
)

var _ = Describe("FromError", func() {
	It("Extracts the status from an status error", func() {
		var err error = &status.Status{
			Code:    codes.OK,
			Message: "Not an error",
		}
		Expect(status.FromError(err)).To(Equal(err.(*status.Status)))
	})

	Context("when the error is not a status", func() {
		It("returns a status with unknown code", func() {
			err := errors.New("Not a status")
			st := status.FromError(err)
			Expect(st.Code).To(Equal(codes.Unknown))
			Expect(st.Message).To(Equal(err.Error()))
		})
	})
})

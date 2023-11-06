package main

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"

	fnv1beta1 "github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/response"

	"github.com/Piotr1215/function-simple-label/label/v1beta1"
)

// Function returns whatever response you ask it to.
type Function struct {
	fnv1beta1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
}

// RunFunction runs the Function.
func (f *Function) RunFunction(_ context.Context, req *fnv1beta1.RunFunctionRequest) (*fnv1beta1.RunFunctionResponse, error) {
	f.log.Info("Running function", "tag", req.GetMeta().GetTag())

	rsp := response.To(req, response.DefaultTTL)

	in := &v1beta1.Input{}
	f.log.Debug("Getting input", "input", in)
	if err := request.GetInput(req, in); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get Function input from %T", req))
		return rsp, nil
	}
	desired, err := request.GetDesiredComposedResources(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get desired composed resources from %T", req))
		return rsp, nil
	}
	f.log.Debug("Found desired resources", "count", len(desired))
	// This will became the value of label "crossplane.io/test.label"
	// If the lable is missing it will be added with the value of label field
	// If the label is present its value will be updated
	for _, dr := range desired {
		if _, ok := dr.Resource.GetLabels()["crossplane.io/test-label"]; ok {
			continue
		}

		meta.AddLabels(dr.Resource, map[string]string{"crossplane.io/test-label": in.Label})
	}
	response.Normalf(rsp, "I was run with input %q!", in.Label)
	f.log.Info("I was run!", "input", in.Label)
	if err := response.SetDesiredComposedResources(rsp, desired); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot set desired composed resources from %T", req))
		return rsp, nil
	}
	return rsp, nil
}

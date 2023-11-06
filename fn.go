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

	// Create a response to the request. This copies the desired state and
	// pipeline context from the request to the response.
	rsp := response.To(req, response.DefaultTTL)

	// This input comes from the label folder in
	// "github.com/Piotr1215/function-simple-label/label/v1beta1"
	in := &v1beta1.Input{}

	// Confirm we are getting input from the request
	f.log.Debug("Getting input", "input", in)
	if err := request.GetInput(req, in); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get Function input from %T", req))
		return rsp, nil
	}

	// Read the observed XR from the request. Most functions use the observed XR
	// to add desired managed resources.
	xr, err := request.GetObservedCompositeResource(req)
	f.log.Debug("Getting XR", "XR", xr)

	if err != nil {

		// If the function can't read the XR, the request is malformed. This
		// should never happen. The function returns a fatal result. This tells
		// Crossplane to stop running functions and return an error.
		response.Fatal(rsp, errors.Wrapf(err, "cannot get observed composite resource from %T", req))
		return rsp, nil

	}

	// Get all desired composed resources from the request. The function will
	// update this map of resources, then save it. This get, update, set pattern
	// ensures the function keeps any resources added by other functions.
	desired, err := request.GetDesiredComposedResources(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get desired composed resources from %T", req))
		return rsp, nil
	}

	f.log.Debug("Found desired resources", "count", len(desired))

	// Main logic of the function
	// Create a label on all desired resources
	// If the lable is missing it will be added with the value of label field
	// If the label is present its value will be updated
	for _, dr := range desired {
		if _, ok := dr.Resource.GetLabels()["crossplane.io/test-label"]; ok {
			continue
		}

		meta.AddLabels(dr.Resource, map[string]string{"crossplane.io/test-label": in.Label})
	}
	response.Normalf(rsp, "I was run with input %q", in.Label)
	f.log.Info("I was run!", "input", in.Label)
	return rsp, nil
}

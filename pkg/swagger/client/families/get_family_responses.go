// Code generated by go-swagger; DO NOT EDIT.

package families

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/fugue/regula/pkg/swagger/models"
)

// GetFamilyReader is a Reader for the GetFamily structure.
type GetFamilyReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetFamilyReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetFamilyOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewGetFamilyBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewGetFamilyUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewGetFamilyForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGetFamilyNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetFamilyInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewGetFamilyOK creates a GetFamilyOK with default headers values
func NewGetFamilyOK() *GetFamilyOK {
	return &GetFamilyOK{}
}

/*GetFamilyOK handles this case with default header values.

The desired Family.
*/
type GetFamilyOK struct {
	Payload *models.FamilyWithRules
}

func (o *GetFamilyOK) Error() string {
	return fmt.Sprintf("[GET /families/{family_id}][%d] getFamilyOK  %+v", 200, o.Payload)
}

func (o *GetFamilyOK) GetPayload() *models.FamilyWithRules {
	return o.Payload
}

func (o *GetFamilyOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.FamilyWithRules)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetFamilyBadRequest creates a GetFamilyBadRequest with default headers values
func NewGetFamilyBadRequest() *GetFamilyBadRequest {
	return &GetFamilyBadRequest{}
}

/*GetFamilyBadRequest handles this case with default header values.

BadRequestError
*/
type GetFamilyBadRequest struct {
	Payload *models.BadRequestError
}

func (o *GetFamilyBadRequest) Error() string {
	return fmt.Sprintf("[GET /families/{family_id}][%d] getFamilyBadRequest  %+v", 400, o.Payload)
}

func (o *GetFamilyBadRequest) GetPayload() *models.BadRequestError {
	return o.Payload
}

func (o *GetFamilyBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.BadRequestError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetFamilyUnauthorized creates a GetFamilyUnauthorized with default headers values
func NewGetFamilyUnauthorized() *GetFamilyUnauthorized {
	return &GetFamilyUnauthorized{}
}

/*GetFamilyUnauthorized handles this case with default header values.

AuthenticationError
*/
type GetFamilyUnauthorized struct {
	Payload *models.AuthenticationError
}

func (o *GetFamilyUnauthorized) Error() string {
	return fmt.Sprintf("[GET /families/{family_id}][%d] getFamilyUnauthorized  %+v", 401, o.Payload)
}

func (o *GetFamilyUnauthorized) GetPayload() *models.AuthenticationError {
	return o.Payload
}

func (o *GetFamilyUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.AuthenticationError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetFamilyForbidden creates a GetFamilyForbidden with default headers values
func NewGetFamilyForbidden() *GetFamilyForbidden {
	return &GetFamilyForbidden{}
}

/*GetFamilyForbidden handles this case with default header values.

AuthorizationError
*/
type GetFamilyForbidden struct {
	Payload *models.AuthorizationError
}

func (o *GetFamilyForbidden) Error() string {
	return fmt.Sprintf("[GET /families/{family_id}][%d] getFamilyForbidden  %+v", 403, o.Payload)
}

func (o *GetFamilyForbidden) GetPayload() *models.AuthorizationError {
	return o.Payload
}

func (o *GetFamilyForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.AuthorizationError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetFamilyNotFound creates a GetFamilyNotFound with default headers values
func NewGetFamilyNotFound() *GetFamilyNotFound {
	return &GetFamilyNotFound{}
}

/*GetFamilyNotFound handles this case with default header values.

NotFoundError
*/
type GetFamilyNotFound struct {
	Payload *models.NotFoundError
}

func (o *GetFamilyNotFound) Error() string {
	return fmt.Sprintf("[GET /families/{family_id}][%d] getFamilyNotFound  %+v", 404, o.Payload)
}

func (o *GetFamilyNotFound) GetPayload() *models.NotFoundError {
	return o.Payload
}

func (o *GetFamilyNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.NotFoundError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetFamilyInternalServerError creates a GetFamilyInternalServerError with default headers values
func NewGetFamilyInternalServerError() *GetFamilyInternalServerError {
	return &GetFamilyInternalServerError{}
}

/*GetFamilyInternalServerError handles this case with default header values.

InternalServerError
*/
type GetFamilyInternalServerError struct {
	Payload *models.InternalServerError
}

func (o *GetFamilyInternalServerError) Error() string {
	return fmt.Sprintf("[GET /families/{family_id}][%d] getFamilyInternalServerError  %+v", 500, o.Payload)
}

func (o *GetFamilyInternalServerError) GetPayload() *models.InternalServerError {
	return o.Payload
}

func (o *GetFamilyInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.InternalServerError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
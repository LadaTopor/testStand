package helper

import (
	"slices"

	"github.com/shopspring/decimal"
)

type GatewayMethodFieldValue struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	GtwId string `json:"gtw_id"`
}

type GatewayMethodField struct {
	Id     string                    `json:"id"`
	Values []GatewayMethodFieldValue `json:"values"`
}

type GatewayMethod struct {
	Id             string               `json:"id"`
	GtwId          string               `json:"gtw_id"`
	MapId          string               `json:"map_id"`
	MapInputId     string               `json:"map_input_id"`
	MapOutputId    string               `json:"map_output_id"`
	RequiredFields []GatewayMethodField `json:"required_fields"`
}

type GatewayTransport struct {
	BaseAddress      string `json:"base_address"`
	SkipVerification bool   `json:"skip_verification"`
	Timeout          *int   `json:"timeout"`
}

type GatewayParams struct {
	Transport            GatewayTransport  `json:"transport"`
	PercentageDifference *decimal.Decimal  `json:"percentage_difference"`
	PaymentMethods       []GatewayMethod   `json:"payment_methods"`
	PayoutMethods        []GatewayMethod   `json:"payout_methods"`
	InBanks              map[string]string `json:"in_banks"`
	OutBanks             map[string]string `json:"out_banks"`
}

// GatewayMethodGetByInternalId, returns 'nil' if not found or argument is zerolen
func GatewayMethodGetByInternalId(methods []GatewayMethod, methodInternalId string) *GatewayMethod {
	if len(methodInternalId) == 0 || len(methods) == 0 {
		return nil
	}

	methodIdx := slices.IndexFunc(methods, func(method GatewayMethod) bool {
		return method.Id == methodInternalId
	})

	if methodIdx < 0 {
		return nil
	}

	return &methods[methodIdx]
}

// GatewayFieldGetByInternalId, returns 'nil' if not found or argument is zerolen
func GatewayFieldGetByInternalId(fields []GatewayMethodField, fieldInternalId string) *GatewayMethodField {
	if len(fields) == 0 || len(fieldInternalId) == 0 {
		return nil
	}

	fieldIdx := slices.IndexFunc(fields, func(field GatewayMethodField) bool {
		return field.Id == fieldInternalId
	})

	if fieldIdx < 0 {
		return nil
	}

	return &fields[fieldIdx]
}

// GatewayFieldValueGetByInternalId, returns 'nil' if not found or argument is zerolen
func GatewayFieldValueGetByInternalId(values []GatewayMethodFieldValue, valueInternalId string) *GatewayMethodFieldValue {
	if len(values) == 0 || len(valueInternalId) == 0 {
		return nil
	}

	valueIdx := slices.IndexFunc(values, func(value GatewayMethodFieldValue) bool {
		return value.Id == valueInternalId
	})

	if valueIdx < 0 {
		return nil
	}

	return &values[valueIdx]
}

// GatewayFieldValueGetByExternalId, returns 'nil' if not found or argument is zerolen
func GatewayFieldValueGetByExternalId(values []GatewayMethodFieldValue, valueExternalId string) *GatewayMethodFieldValue {
	if len(values) == 0 || len(valueExternalId) == 0 {
		return nil
	}

	valueIdx := slices.IndexFunc(values, func(value GatewayMethodFieldValue) bool {
		return value.GtwId == valueExternalId
	})

	if valueIdx < 0 {
		return nil
	}

	return &values[valueIdx]
}

func GatewayFieldValueGetByInternalPathAndExternalValueId(methods []GatewayMethod, methodInternalId string, fieldInternalId string, valueExternalId string) *GatewayMethodFieldValue {
	if len(methods) == 0 {
		return nil
	}

	method := GatewayMethodGetByInternalId(methods, methodInternalId)
	if method == nil {
		return nil
	}

	// TODO: add optional fields in future
	methodField := GatewayFieldGetByInternalId(method.RequiredFields, fieldInternalId)
	if methodField == nil {
		return nil
	}

	methodFieldValue := GatewayFieldValueGetByExternalId(methodField.Values, fieldInternalId)
	if methodFieldValue == nil {
		return nil
	}

	return methodFieldValue
}

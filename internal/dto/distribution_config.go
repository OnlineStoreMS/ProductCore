package dto

type DistributionConfigOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type DistributionConfigMeta struct {
	ShipModes         []DistributionConfigOption `json:"shipModes"`
	SpotShipTimes     []DistributionConfigOption `json:"spotShipTimes"`
	AfterSalePolicies []DistributionConfigOption `json:"afterSalePolicies"`
}

type DistributionConfigDTO struct {
	ID                   uint64 `json:"id,omitempty"`
	PlatformTypeID       uint64 `json:"platformTypeId"`
	PlatformTypeCode     string `json:"platformTypeCode,omitempty"`
	PlatformTypeName     string `json:"platformTypeName,omitempty"`
	ShipMode             string `json:"shipMode"`
	ShipModeLabel        string `json:"shipModeLabel,omitempty"`
	SpotShipTime         string `json:"spotShipTime"`
	SpotShipTimeLabel    string `json:"spotShipTimeLabel,omitempty"`
	FreightTemplate      string `json:"freightTemplate"`
	AfterSalePolicy      string `json:"afterSalePolicy"`
	AfterSalePolicyLabel string `json:"afterSalePolicyLabel,omitempty"`
	Enabled              int8   `json:"enabled"`
	Remark               string `json:"remark"`
}

func DistributionConfigOptions() DistributionConfigMeta {
	return DistributionConfigMeta{
		ShipModes: []DistributionConfigOption{
			{Value: "spot", Label: "现货发货"},
			{Value: "presale", Label: "预售发货"},
		},
		SpotShipTimes: []DistributionConfigOption{
			{Value: "same_day", Label: "当日发"},
			{Value: "next_day", Label: "次日发"},
			{Value: "hours_48", Label: "48小时"},
		},
		AfterSalePolicies: []DistributionConfigOption{
			{Value: "7day", Label: "7天无理由退货"},
			{Value: "7day_custom", Label: "7天无理由退货（定制类不支持）"},
			{Value: "none", Label: "不支持7天无理由"},
		},
	}
}

func DistributionConfigLabel(options []DistributionConfigOption, value string) string {
	for _, opt := range options {
		if opt.Value == value {
			return opt.Label
		}
	}
	return value
}

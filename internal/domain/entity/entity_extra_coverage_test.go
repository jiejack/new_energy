package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCarbonEmissionFactor_TableName_Extra(t *testing.T) {
	f := &CarbonEmissionFactor{}
	assert.Equal(t, "carbon_emission_factors", f.TableName())
}

func TestCarbonEmissionRecord_TableName_Extra(t *testing.T) {
	r := &CarbonEmissionRecord{}
	assert.Equal(t, "carbon_emission_records", r.TableName())
}

func TestCarbonEmissionSummary_TableName_Extra(t *testing.T) {
	s := &CarbonEmissionSummary{}
	assert.Equal(t, "carbon_emission_summaries", s.TableName())
}

func TestCarbonReductionTarget_TableName_Extra(t *testing.T) {
	tr := &CarbonReductionTarget{}
	assert.Equal(t, "carbon_reduction_targets", tr.TableName())
}

func TestNewCarbonEmissionFactor_Extra(t *testing.T) {
	effectiveAt := time.Now()
	f := NewCarbonEmissionFactor("电力排放因子", "EF-001", CarbonEmissionScope2, "IPCC", 0.5810, "tCO2/MWh", "v1.0", effectiveAt)
	assert.Equal(t, "电力排放因子", f.Name)
	assert.Equal(t, "EF-001", f.Code)
	assert.Equal(t, CarbonEmissionScope2, f.Scope)
	assert.Equal(t, "IPCC", f.Source)
	assert.InDelta(t, 0.5810, f.Value, 0.0001)
	assert.Equal(t, "tCO2/MWh", f.Unit)
	assert.Equal(t, "v1.0", f.Version)
	assert.Equal(t, effectiveAt, f.EffectiveAt)
	assert.True(t, f.IsActive)
	assert.False(t, f.CreatedAt.IsZero())
	assert.False(t, f.UpdatedAt.IsZero())
}

func TestNewCarbonEmissionRecord_Extra(t *testing.T) {
	recordTime := time.Now()
	r := NewCarbonEmissionRecord(recordTime, CarbonEmissionScope1, "target-1", "目标站", "factor-1", "FC-001", 0.5, 100.0, "MWh", "2024-01")
	assert.Equal(t, recordTime, r.RecordTime)
	assert.Equal(t, CarbonEmissionScope1, r.Scope)
	assert.Equal(t, "target-1", r.TargetID)
	assert.Equal(t, "目标站", r.TargetName)
	assert.Equal(t, "factor-1", r.FactorID)
	assert.Equal(t, "FC-001", r.FactorCode)
	assert.InDelta(t, 0.5, r.FactorValue, 0.0001)
	assert.InDelta(t, 100.0, r.ActivityData, 0.0001)
	assert.Equal(t, "MWh", r.ActivityUnit)
	assert.InDelta(t, 50.0, r.EmissionValue, 0.0001)
	assert.Equal(t, "tCO2e", r.EmissionUnit)
	assert.Equal(t, "2024-01", r.Period)
	assert.Equal(t, CarbonEmissionStatusDraft, r.Status)
	assert.False(t, r.CreatedAt.IsZero())
	assert.False(t, r.UpdatedAt.IsZero())
}

func TestCarbonEmissionScope_Constants_Extra(t *testing.T) {
	assert.Equal(t, CarbonEmissionScope("scope1"), CarbonEmissionScope1)
	assert.Equal(t, CarbonEmissionScope("scope2"), CarbonEmissionScope2)
	assert.Equal(t, CarbonEmissionScope("scope3"), CarbonEmissionScope3)
}

func TestCarbonEmissionStatus_Constants_Extra(t *testing.T) {
	assert.Equal(t, CarbonEmissionStatus("draft"), CarbonEmissionStatusDraft)
	assert.Equal(t, CarbonEmissionStatus("pending"), CarbonEmissionStatusPending)
	assert.Equal(t, CarbonEmissionStatus("approved"), CarbonEmissionStatusApproved)
	assert.Equal(t, CarbonEmissionStatus("rejected"), CarbonEmissionStatusRejected)
}

func TestConfigItem_TableName_Extra(t *testing.T) {
	c := &ConfigItem{}
	assert.Equal(t, "config_items", c.TableName())
}

func TestConfigVersion_TableName_Extra(t *testing.T) {
	cv := &ConfigVersion{}
	assert.Equal(t, "config_versions", cv.TableName())
}

func TestConfigRelease_TableName_Extra(t *testing.T) {
	cr := &ConfigRelease{}
	assert.Equal(t, "config_releases", cr.TableName())
}

func TestConfigAudit_TableName_Extra(t *testing.T) {
	ca := &ConfigAudit{}
	assert.Equal(t, "config_audits", ca.TableName())
}

func TestMaintenanceRecord_TableName_Extra(t *testing.T) {
	m := &MaintenanceRecord{}
	assert.Equal(t, "maintenance_records", m.TableName())
}

func TestSparePart_TableName_Extra(t *testing.T) {
	s := &SparePart{}
	assert.Equal(t, "spare_parts", s.TableName())
}

func TestDeviceDocument_TableName_Extra(t *testing.T) {
	d := &DeviceDocument{}
	assert.Equal(t, "device_documents", d.TableName())
}

func TestWorkOrder_TableName_Extra(t *testing.T) {
	w := &WorkOrder{}
	assert.Equal(t, "work_orders", w.TableName())
}

func TestInventory_TableName_Extra(t *testing.T) {
	i := &Inventory{}
	assert.Equal(t, "inventory", i.TableName())
}

func TestInventoryTransaction_TableName_Extra(t *testing.T) {
	it := &InventoryTransaction{}
	assert.Equal(t, "inventory_transactions", it.TableName())
}

func TestSupplier_TableName_Extra(t *testing.T) {
	s := &Supplier{}
	assert.Equal(t, "suppliers", s.TableName())
}

func TestPurchaseOrder_TableName_Extra(t *testing.T) {
	po := &PurchaseOrder{}
	assert.Equal(t, "purchase_orders", po.TableName())
}

func TestPurchaseOrderItem_TableName_Extra(t *testing.T) {
	poi := &PurchaseOrderItem{}
	assert.Equal(t, "purchase_order_items", poi.TableName())
}

func TestReceipt_TableName_Extra(t *testing.T) {
	r := &Receipt{}
	assert.Equal(t, "receipts", r.TableName())
}

func TestReceiptItem_TableName_Extra(t *testing.T) {
	ri := &ReceiptItem{}
	assert.Equal(t, "receipt_items", ri.TableName())
}

func TestCostCategory_TableName_Extra(t *testing.T) {
	cc := &CostCategory{}
	assert.Equal(t, "cost_categories", cc.TableName())
}

func TestCostEntry_TableName_Extra(t *testing.T) {
	ce := &CostEntry{}
	assert.Equal(t, "cost_entries", ce.TableName())
}

func TestCostAllocation_TableName_Extra(t *testing.T) {
	ca := &CostAllocation{}
	assert.Equal(t, "cost_allocations", ca.TableName())
}

func TestCostReport_TableName_Extra(t *testing.T) {
	cr := &CostReport{}
	assert.Equal(t, "cost_reports", cr.TableName())
}

func TestAsset_TableName_Extra(t *testing.T) {
	a := &Asset{}
	assert.Equal(t, "assets", a.TableName())
}

func TestAssetMaintenanceRecord_TableName_Extra(t *testing.T) {
	amr := &AssetMaintenanceRecord{}
	assert.Equal(t, "asset_maintenance_records", amr.TableName())
}

func TestAssetDepreciationRecord_TableName_Extra(t *testing.T) {
	adr := &AssetDepreciationRecord{}
	assert.Equal(t, "asset_depreciation_records", adr.TableName())
}

func TestAssetDocument_TableName_Extra(t *testing.T) {
	ad := &AssetDocument{}
	assert.Equal(t, "asset_documents", ad.TableName())
}

func TestEdgeNode_TableName_Extra(t *testing.T) {
	e := EdgeNode{}
	assert.Equal(t, "edge_nodes", e.TableName())
}

func TestEdgeNodeStatus_Constants_Extra(t *testing.T) {
	assert.Equal(t, EdgeNodeStatus("online"), EdgeNodeOnline)
	assert.Equal(t, EdgeNodeStatus("offline"), EdgeNodeOffline)
	assert.Equal(t, EdgeNodeStatus("degraded"), EdgeNodeDegraded)
	assert.Equal(t, EdgeNodeStatus("maintenance"), EdgeNodeMaintenance)
}

func TestEnergyEfficiencyRecord_TableName_Extra(t *testing.T) {
	r := &EnergyEfficiencyRecord{}
	assert.Equal(t, "energy_efficiency_records", r.TableName())
}

func TestEnergyEfficiencyAnalysis_TableName_Extra(t *testing.T) {
	a := &EnergyEfficiencyAnalysis{}
	assert.Equal(t, "energy_efficiency_analyses", a.TableName())
}

func TestNewEnergyEfficiencyRecord_Extra(t *testing.T) {
	now := time.Now()

	t.Run("normal_efficiency", func(t *testing.T) {
		r := NewEnergyEfficiencyRecord(now, EnergyEfficiencyTypeDevice, "dev-1", "设备1", 100.0, 80.0, "2024-01")
		assert.Equal(t, now, r.RecordTime)
		assert.Equal(t, EnergyEfficiencyTypeDevice, r.Type)
		assert.Equal(t, "dev-1", r.TargetID)
		assert.Equal(t, "设备1", r.TargetName)
		assert.InDelta(t, 100.0, r.InputEnergy, 0.0001)
		assert.InDelta(t, 80.0, r.OutputEnergy, 0.0001)
		assert.InDelta(t, 0.8, r.Efficiency, 0.0001)
		assert.Equal(t, EnergyEfficiencyLevelNormal, r.EfficiencyLevel)
		assert.Equal(t, "2024-01", r.Period)
	})

	t.Run("excellent_efficiency", func(t *testing.T) {
		r := NewEnergyEfficiencyRecord(now, EnergyEfficiencyTypeStation, "st-1", "站点1", 100.0, 96.0, "2024-01")
		assert.InDelta(t, 0.96, r.Efficiency, 0.0001)
		assert.Equal(t, EnergyEfficiencyLevelExcellent, r.EfficiencyLevel)
	})

	t.Run("good_efficiency", func(t *testing.T) {
		r := NewEnergyEfficiencyRecord(now, EnergyEfficiencyTypeSystem, "sys-1", "系统1", 100.0, 88.0, "2024-01")
		assert.InDelta(t, 0.88, r.Efficiency, 0.0001)
		assert.Equal(t, EnergyEfficiencyLevelGood, r.EfficiencyLevel)
	})

	t.Run("poor_efficiency", func(t *testing.T) {
		r := NewEnergyEfficiencyRecord(now, EnergyEfficiencyTypeComprehensive, "comp-1", "综合1", 100.0, 60.0, "2024-01")
		assert.InDelta(t, 0.6, r.Efficiency, 0.0001)
		assert.Equal(t, EnergyEfficiencyLevelPoor, r.EfficiencyLevel)
	})

	t.Run("zero_input", func(t *testing.T) {
		r := NewEnergyEfficiencyRecord(now, EnergyEfficiencyTypeDevice, "dev-2", "设备2", 0.0, 0.0, "2024-01")
		assert.InDelta(t, 0.0, r.Efficiency, 0.0001)
		assert.Equal(t, EnergyEfficiencyLevelPoor, r.EfficiencyLevel)
	})
}

func TestEnergyEfficiencyType_Constants_Extra(t *testing.T) {
	assert.Equal(t, EnergyEfficiencyType("device"), EnergyEfficiencyTypeDevice)
	assert.Equal(t, EnergyEfficiencyType("station"), EnergyEfficiencyTypeStation)
	assert.Equal(t, EnergyEfficiencyType("system"), EnergyEfficiencyTypeSystem)
	assert.Equal(t, EnergyEfficiencyType("comprehensive"), EnergyEfficiencyTypeComprehensive)
}

func TestEnergyEfficiencyLevel_Constants_Extra(t *testing.T) {
	assert.Equal(t, EnergyEfficiencyLevel("excellent"), EnergyEfficiencyLevelExcellent)
	assert.Equal(t, EnergyEfficiencyLevel("good"), EnergyEfficiencyLevelGood)
	assert.Equal(t, EnergyEfficiencyLevel("normal"), EnergyEfficiencyLevelNormal)
	assert.Equal(t, EnergyEfficiencyLevel("poor"), EnergyEfficiencyLevelPoor)
}

func TestFaultDetectionResult_TableName_Extra(t *testing.T) {
	f := FaultDetectionResult{}
	assert.Equal(t, "fault_detection_results", f.TableName())
}

func TestFaultDetectionResult_Confirm_Extra(t *testing.T) {
	f := NewFaultDetectionResult("dev-1", "overheating", FaultSeverityCritical, 0.95, "温度过高", "v1.0")
	assert.Equal(t, FaultDetectionStatusPending, f.Status)
	f.Confirm()
	assert.Equal(t, FaultDetectionStatusConfirmed, f.Status)
}

func TestFaultDetectionResult_Resolve_Extra(t *testing.T) {
	f := NewFaultDetectionResult("dev-1", "overheating", FaultSeverityCritical, 0.95, "温度过高", "v1.0")
	f.Resolve()
	assert.Equal(t, FaultDetectionStatusResolved, f.Status)
}

func TestFaultDetectionResult_Ignore_Extra(t *testing.T) {
	f := NewFaultDetectionResult("dev-1", "overheating", FaultSeverityCritical, 0.95, "温度过高", "v1.0")
	f.Ignore()
	assert.Equal(t, FaultDetectionStatusIgnored, f.Status)
}

func TestFaultDetectionResult_SetRootCause_Extra(t *testing.T) {
	f := NewFaultDetectionResult("dev-1", "overheating", FaultSeverityCritical, 0.95, "温度过高", "v1.0")
	assert.Nil(t, f.RootCause)
	f.SetRootCause("散热风扇故障")
	assert.NotNil(t, f.RootCause)
	assert.Equal(t, "散热风扇故障", *f.RootCause)
}

func TestFaultDetectionResult_SetRecommendation_Extra(t *testing.T) {
	f := NewFaultDetectionResult("dev-1", "overheating", FaultSeverityCritical, 0.95, "温度过高", "v1.0")
	assert.Nil(t, f.Recommendation)
	f.SetRecommendation("更换散热风扇")
	assert.NotNil(t, f.Recommendation)
	assert.Equal(t, "更换散热风扇", *f.Recommendation)
}

func TestFaultDetectionResult_SetRUL_Extra(t *testing.T) {
	f := NewFaultDetectionResult("dev-1", "overheating", FaultSeverityCritical, 0.95, "温度过高", "v1.0")
	assert.Nil(t, f.RemainingUsefulLifeHrs)
	f.SetRUL(720)
	assert.NotNil(t, f.RemainingUsefulLifeHrs)
	assert.Equal(t, 720, *f.RemainingUsefulLifeHrs)
}

func TestFaultDetectionResult_SetHealthScore_Extra(t *testing.T) {
	f := NewFaultDetectionResult("dev-1", "overheating", FaultSeverityCritical, 0.95, "温度过高", "v1.0")
	assert.Nil(t, f.HealthScore)
	f.SetHealthScore(65)
	assert.NotNil(t, f.HealthScore)
	assert.Equal(t, 65, *f.HealthScore)
}

func TestFaultDetectionResult_LinkWorkOrder_Extra(t *testing.T) {
	f := NewFaultDetectionResult("dev-1", "overheating", FaultSeverityCritical, 0.95, "温度过高", "v1.0")
	assert.Nil(t, f.WorkOrderID)
	f.LinkWorkOrder("wo-123")
	assert.NotNil(t, f.WorkOrderID)
	assert.Equal(t, "wo-123", *f.WorkOrderID)
}

func TestFaultSeverity_Constants_Extra(t *testing.T) {
	assert.Equal(t, FaultSeverity("info"), FaultSeverityInfo)
	assert.Equal(t, FaultSeverity("warning"), FaultSeverityWarning)
	assert.Equal(t, FaultSeverity("critical"), FaultSeverityCritical)
	assert.Equal(t, FaultSeverity("fatal"), FaultSeverityFatal)
}

func TestFaultDetectionStatus_Constants_Extra(t *testing.T) {
	assert.Equal(t, FaultDetectionStatus(1), FaultDetectionStatusPending)
	assert.Equal(t, FaultDetectionStatus(2), FaultDetectionStatusConfirmed)
	assert.Equal(t, FaultDetectionStatus(3), FaultDetectionStatusResolved)
	assert.Equal(t, FaultDetectionStatus(4), FaultDetectionStatusIgnored)
}

func TestForecastResult_TableName_Extra(t *testing.T) {
	f := ForecastResult{}
	assert.Equal(t, "forecast_results", f.TableName())
}

func TestForecastResult_SetActualPower_Extra(t *testing.T) {
	f := NewForecastResult("station-1", ForecastTypeShortTerm, time.Now(), 100.5, "v2.0")
	assert.Nil(t, f.ActualPower)
	f.SetActualPower(98.3)
	assert.NotNil(t, f.ActualPower)
	assert.InDelta(t, 98.3, *f.ActualPower, 0.0001)
}

func TestForecastResult_SetAccuracy_Extra(t *testing.T) {
	f := NewForecastResult("station-1", ForecastTypeShortTerm, time.Now(), 100.5, "v2.0")
	assert.Nil(t, f.Accuracy)
	f.SetAccuracy(0.95)
	assert.NotNil(t, f.Accuracy)
	assert.InDelta(t, 0.95, *f.Accuracy, 0.0001)
}

func TestForecastResult_SetConfidenceInterval_Extra(t *testing.T) {
	f := NewForecastResult("station-1", ForecastTypeShortTerm, time.Now(), 100.5, "v2.0")
	assert.Nil(t, f.ConfidenceLower)
	assert.Nil(t, f.ConfidenceUpper)
	f.SetConfidenceInterval(90.0, 110.0)
	assert.NotNil(t, f.ConfidenceLower)
	assert.NotNil(t, f.ConfidenceUpper)
	assert.InDelta(t, 90.0, *f.ConfidenceLower, 0.0001)
	assert.InDelta(t, 110.0, *f.ConfidenceUpper, 0.0001)
}

func TestForecastResult_SetAttribution_Extra(t *testing.T) {
	f := NewForecastResult("station-1", ForecastTypeShortTerm, time.Now(), 100.5, "v2.0")
	assert.Nil(t, f.AttributionType)
	assert.Nil(t, f.AttributionDetail)
	f.SetAttribution("weather", "云量增加导致发电量下降")
	assert.NotNil(t, f.AttributionType)
	assert.NotNil(t, f.AttributionDetail)
	assert.Equal(t, "weather", *f.AttributionType)
	assert.Equal(t, "云量增加导致发电量下降", *f.AttributionDetail)
}

func TestForecastType_Constants_Extra(t *testing.T) {
	assert.Equal(t, ForecastType("ultra_short_term"), ForecastTypeUltraShortTerm)
	assert.Equal(t, ForecastType("short_term"), ForecastTypeShortTerm)
	assert.Equal(t, ForecastType("medium_term"), ForecastTypeMediumTerm)
}

func TestModelVersion_TableName_Extra(t *testing.T) {
	mv := ModelVersion{}
	assert.Equal(t, "model_versions", mv.TableName())
}

func TestModelStatus_Constants_Extra(t *testing.T) {
	assert.Equal(t, ModelStatus("training"), ModelStatusTraining)
	assert.Equal(t, ModelStatus("staging"), ModelStatusStaging)
	assert.Equal(t, ModelStatus("production"), ModelStatusProd)
	assert.Equal(t, ModelStatus("retired"), ModelStatusRetired)
}

func TestNewFaultDetectionResult_Extra(t *testing.T) {
	f := NewFaultDetectionResult("dev-1", "vibration", FaultSeverityWarning, 0.85, "异常振动", "v1.5")
	assert.Equal(t, "dev-1", f.DeviceID)
	assert.Equal(t, "vibration", f.FaultType)
	assert.Equal(t, FaultSeverityWarning, f.Severity)
	assert.InDelta(t, 0.85, f.Confidence, 0.0001)
	assert.Equal(t, "异常振动", f.Description)
	assert.Equal(t, FaultDetectionStatusPending, f.Status)
	assert.Equal(t, "v1.5", f.ModelVersion)
}

func TestNewForecastResult_Extra(t *testing.T) {
	targetTime := time.Now().Add(24 * time.Hour)
	f := NewForecastResult("station-1", ForecastTypeMediumTerm, targetTime, 200.0, "v3.0")
	assert.Equal(t, "station-1", f.StationID)
	assert.Equal(t, ForecastTypeMediumTerm, f.ForecastType)
	assert.Equal(t, targetTime, f.TargetTime)
	assert.InDelta(t, 200.0, f.PredictedPower, 0.0001)
	assert.Equal(t, "v3.0", f.ModelVersion)
	assert.Nil(t, f.ActualPower)
	assert.Nil(t, f.Accuracy)
}

func TestMapJSON_Extra(t *testing.T) {
	m := MapJSON{"key1": "value1", "key2": "value2"}
	assert.Equal(t, "value1", m["key1"])
	assert.Equal(t, "value2", m["key2"])
}

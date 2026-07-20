package inspection

import "testing"

func TestTableNames(t *testing.T) {
	cases := map[string]string{
		InspCluster{}.TableName():          "insp_clusters",
		InspRule{}.TableName():             "insp_rules",
		InspTask{}.TableName():             "insp_tasks",
		InspInspection{}.TableName():       "insp_inspections",
		InspInspectionDetail{}.TableName(): "insp_inspection_details",
		InspAlert{}.TableName():            "insp_alerts",
		InspReport{}.TableName():           "insp_reports",
	}
	for got, want := range cases {
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	}
}

package learning

import (
	"testing"
	"time"
)

func TestEligibleActivitiesRespectsPrerequisitesAndReviews(t *testing.T) {
	m := Manifest{Milestones: []Milestone{{Nodes: []Node{
		{ID: "start", SkillID: "start"},
		{ID: "next", SkillID: "next", Prerequisites: []string{"start"}},
		{ID: "later", SkillID: "later", Prerequisites: []string{"next"}},
	}}}}
	choices := eligibleActivities(m, map[string]bool{}, nil)
	if len(choices) != 1 || choices[0].Node.ID != "start" {
		t.Fatalf("initial choices: %+v", choices)
	}
	choices = eligibleActivities(m, map[string]bool{}, []activityChoice{{Node: m.Milestones[0].Nodes[2], DueAt: time.Now()}})
	if len(choices) != 1 || choices[0].Node.ID != "start" {
		t.Fatalf("due review bypassed prerequisites: %+v", choices)
	}
	choices = eligibleActivities(m, map[string]bool{"start": true}, []activityChoice{{Node: m.Milestones[0].Nodes[0], DueAt: time.Now()}})
	if len(choices) != 2 || choices[0].Node.ID != "start" || choices[1].Node.ID != "next" {
		t.Fatalf("review and unlocked choices: %+v", choices)
	}
	choices = eligibleActivities(m, map[string]bool{"start": true}, []activityChoice{{Node: m.Milestones[0].Nodes[1], DueAt: time.Now()}})
	if len(choices) != 1 || choices[0].Node.ID != "next" {
		t.Fatalf("due activity was duplicated: %+v", choices)
	}
}

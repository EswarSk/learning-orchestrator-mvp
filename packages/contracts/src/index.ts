import { z } from "zod";

export const LOCAL_AUTH_TOKEN = "local-development";

export const UUID = z.uuid();
export const IsoInstant = z.iso.datetime({ offset: true });
export const TimeZone = z.string().min(1).max(64).refine((value) => {
  try { Intl.DateTimeFormat(undefined, { timeZone: value }); return true; } catch { return false; }
}, "Invalid IANA time zone");

export const Assistance = z.enum(["none", "replay", "hint", "model_answer"]);
export const Outcome = z.enum(["observed", "needs_help", "uncertain"]);
export const SessionMode = z.enum(["text", "voice"]);
export const SessionState = z.enum(["reserved", "starting", "active", "paused", "assessing", "completed", "abandoned", "failed", "expired"]);
export const OpportunityState = z.enum(["candidate", "ready", "scheduled", "offered", "started", "completed", "dismissed", "suppressed", "canceled", "expired", "failed"]);

export const ActivityPlan = z.object({
  schemaVersion: z.literal(1),
  skillId: z.string().min(1),
  templateId: z.string().min(1),
  templateVersion: z.int().positive(),
  format: z.enum(["listen_respond", "dialogue", "recall"]),
  difficulty: z.int().min(1).max(5),
  objective: z.string().max(160),
  supportLevel: z.enum(["guided", "hint_available", "independent"]),
  situation: z.enum(["dining", "social", "travel", "daily_routine", "ordinary_practice"]),
  situationConfidence: z.number().min(0).max(1),
  sourceArtifactIds: z.array(UUID).max(25),
  taskIds: z.array(z.string().min(1)).min(1).max(10),
}).strict();

export const AssessmentProposal = z.object({
  schemaVersion: z.literal(1),
  sessionId: UUID,
  observations: z.array(z.object({
    taskId: z.string().min(1),
    skillId: z.string().min(1),
    outcome: Outcome,
    assistance: Assistance,
    confidence: z.number().min(0).max(1),
    citedMessageIds: z.array(z.string().min(1)).max(20),
    correction: z.string().max(400).nullable(),
  }).strict()).min(1).max(10),
  learnerFeedback: z.string().max(600),
}).strict();

export const ContextEvent = z.object({
  schemaVersion: z.literal(1),
  id: UUID,
  userId: UUID,
  source: z.enum(["google_calendar", "ios_calendar", "calendar", "ios_geofence", "learning"]),
  connectionId: UUID.nullable(),
  sourceResourceId: z.string().max(255).nullable(),
  sourceEventId: z.string().min(1).max(255),
  sourceRevision: z.string().min(1).max(255),
  observedAt: IsoInstant,
  receivedAt: IsoInstant,
  validUntil: IsoInstant,
  facts: z.record(z.string(), z.unknown()),
  provenance: z.object({
    sharingClass: z.enum(["private_only", "eligible_first_party"]),
    sourceArtifactIds: z.array(UUID).max(100),
  }).strict(),
}).strict();

export const LocationSignal = z.object({
  schemaVersion: z.literal(1),
  installationId: z.string().min(1).max(128),
  clientEventId: z.string().min(1).max(128),
  regionId: z.string().min(1).max(128),
  regionVersion: z.int().positive(),
  transition: z.enum(["enter", "exit"]),
  observedAt: IsoInstant,
}).strict();

export const SessionSource = z.discriminatedUnion("type", [
  z.object({ type: z.literal("opportunity"), id: UUID }).strict(),
  z.object({ type: z.literal("curriculum"), nodeId: z.string().min(1) }).strict(),
  z.object({ type: z.literal("review"), skillId: z.string().min(1) }).strict(),
]);

export const StartSessionRequest = z.object({ source: SessionSource, mode: SessionMode }).strict();
export const TurnRequest = z.object({
  messageId: z.string().min(1).max(128),
  text: z.string().trim().min(1).max(2_000),
  assistanceRequested: z.enum(["none", "replay", "hint", "model_answer"]).default("none"),
}).strict();
export const FinishSessionRequest = z.object({ reason: z.enum(["completed", "userEnded"]) }).strict();
export const OpportunityAction = z.object({
  action: z.enum(["dismiss", "snooze"]),
  snoozeUntil: IsoInstant.optional(),
}).strict().superRefine((value, ctx) => {
  if (value.action === "snooze" && !value.snoozeUntil) ctx.addIssue({ code: "custom", message: "snoozeUntil is required" });
});

export const PatchMeRequest = z.object({
  goal: z.string().max(160).optional(),
  levelHint: z.enum(["new", "some", "comfortable"]).optional(),
  interests: z.array(z.string().max(50)).max(10).optional(),
  timezone: TimeZone.optional(),
  quietHours: z.object({ start: z.string().regex(/^([01]\d|2[0-3]):[0-5]\d$/), end: z.string().regex(/^([01]\d|2[0-3]):[0-5]\d$/) }).optional(),
  proactivePaused: z.boolean().optional(),
}).strict();

export const ConsentRequest = z.object({
  purpose: z.enum(["calendar_context", "ios_calendar_context", "google_calendar_context", "location_context", "notifications", "shared_patterns"]),
  policyVersion: z.string().min(1).max(32),
  granted: z.boolean(),
}).strict();

export const ErrorEnvelope = z.object({
  error: z.object({ code: z.string(), message: z.string(), retryable: z.boolean(), requestId: z.string() }),
});

export const schemas = {
  ActivityPlan, AssessmentProposal, ContextEvent, LocationSignal, StartSessionRequest,
  TurnRequest, FinishSessionRequest, OpportunityAction, PatchMeRequest, ConsentRequest, ErrorEnvelope,
} as const;

export type ActivityPlan = z.infer<typeof ActivityPlan>;
export type AssessmentProposal = z.infer<typeof AssessmentProposal>;
export type ContextEvent = z.infer<typeof ContextEvent>;
export type LocationSignal = z.infer<typeof LocationSignal>;
export type StartSessionRequest = z.infer<typeof StartSessionRequest>;

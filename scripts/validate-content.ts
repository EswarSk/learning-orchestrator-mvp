import { readFile, readdir } from "node:fs/promises";
import { join } from "node:path";
import { fileURLToPath } from "node:url";
import { z } from "zod";

const text = z.string().trim().min(1);
const Node = z.object({
  id: text, skillId: text, kind: z.enum(["guided", "context", "transfer"]),
  title: text, objective: text, prerequisites: z.array(text), prompt: text,
  acceptedVariation: z.array(text).min(1), commonError: text, support: z.array(text).min(1),
  rubric: text, transferPrompt: text, fallback: text,
}).strict();
const Manifest = z.object({
  schemaVersion: z.literal(1), contentVersion: text,
  subject: z.object({ id: text, title: text, category: text }).strict(),
  reviewStatus: z.enum(["reviewed", "unreviewed"]), reviewNote: text,
  milestones: z.array(z.object({ id: text, title: text, nodes: z.array(Node).min(1) }).strict()).min(1),
}).strict();

const root = fileURLToPath(new URL("../content/", import.meta.url));
const packs = (await readdir(root, { withFileTypes: true })).filter(entry => entry.isDirectory());
if (!packs.length) throw new Error("No content packs found");
let total = 0;
for (const pack of packs) {
  const manifest = Manifest.parse(JSON.parse(await readFile(join(root, pack.name, "manifest.json"), "utf8")));
  if (manifest.subject.id !== pack.name) throw new Error(`${pack.name}: subject ID must match its directory`);
  const nodes = manifest.milestones.flatMap(group => group.nodes);
  const unique = (values: string[], name: string) => {
    if (new Set(values).size !== values.length) throw new Error(`${pack.name}: duplicate ${name}`);
  };
  unique(manifest.milestones.map(group => group.id), "milestone ID");
  unique(nodes.map(node => node.id), "node ID");
  unique(nodes.map(node => node.skillId), "skill ID");
  const bySkill = new Map(nodes.map(node => [node.skillId, node]));
  for (const node of nodes) for (const prerequisite of node.prerequisites) {
    if (!bySkill.has(prerequisite)) throw new Error(`${pack.name}/${node.id}: missing prerequisite ${prerequisite}`);
  }
  const visiting = new Set<string>(); const visited = new Set<string>();
  function visit(skillId: string): void {
    if (visiting.has(skillId)) throw new Error(`${pack.name}: prerequisite cycle at ${skillId}`);
    if (visited.has(skillId)) return;
    visiting.add(skillId);
    for (const prerequisite of bySkill.get(skillId)!.prerequisites) visit(prerequisite);
    visiting.delete(skillId); visited.add(skillId);
  }
  for (const skillId of bySkill.keys()) visit(skillId);
  total += nodes.length;
  console.log(`${pack.name}: ${nodes.length} nodes, ${manifest.reviewStatus}, ${manifest.contentVersion}`);
}
console.log(`validated ${total} nodes across ${packs.length} content pack(s)`);

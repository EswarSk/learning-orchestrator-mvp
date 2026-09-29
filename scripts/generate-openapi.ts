import { writeFile } from "node:fs/promises";
import { schemas } from "../packages/contracts/src/index.js";
import { z } from "zod";

const components = Object.fromEntries(Object.entries(schemas).map(([name, schema]) => [name, z.toJSONSchema(schema)]));
const document = {
  openapi: "3.1.0",
  info: { title: "Learning Orchestrator API", version: "0.1.0" },
  paths: {},
  components: { schemas: components },
};
await writeFile(new URL("../packages/contracts/openapi.json", import.meta.url), `${JSON.stringify(document, null, 2)}\n`);

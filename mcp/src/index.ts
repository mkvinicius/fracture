#!/usr/bin/env node
/**
 * FRACTURE MCP Server
 * Exposes FRACTURE disruption simulation tools to Cursor, VS Code, and Windsurf
 * via the Model Context Protocol (MCP).
 *
 * Tools available:
 *   - fracture_simulate     → Run a full disruption simulation
 *   - fracture_quick_pulse  → Fast 5-second market tension check
 *   - fracture_list_rules   → List the rules of a market domain
 *   - fracture_archetypes   → List available agent archetypes
 *   - fracture_history      → Get past simulation results
 *   - fracture_feedback     → Submit real-world outcome to calibrate archetypes
 */

import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
  Tool,
} from "@modelcontextprotocol/sdk/types.js";

// ─── FRACTURE local server URL (must be running) ─────────────────────────────
const FRACTURE_BASE_URL = process.env.FRACTURE_URL ?? "http://localhost:3000";

// ─── Helper ──────────────────────────────────────────────────────────────────
async function callFracture(path: string, body?: unknown): Promise<unknown> {
  const url = `${FRACTURE_BASE_URL}/api${path}`;
  const res = await fetch(url, {
    method: body ? "POST" : "GET",
    headers: { "Content-Type": "application/json" },
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`FRACTURE API error ${res.status}: ${text}`);
  }
  return res.json();
}

// ─── Tool definitions ─────────────────────────────────────────────────────────
const TOOLS: Tool[] = [
  {
    name: "fracture_simulate",
    description:
      "Run a full FRACTURE disruption simulation. Sends a strategic question to 12 AI agents (8 Conformists + 4 Disruptors) who interact over multiple rounds. Returns: Probable Future, Tension Map, and top Rupture Scenarios — including how YOU could be the one to break the market first.",
    inputSchema: {
      type: "object",
      properties: {
        question: {
          type: "string",
          description:
            "The strategic question to simulate. Examples: 'What happens if our main competitor goes free tomorrow?', 'How would AI replace our sales team in 2 years?'",
        },
        domain: {
          type: "string",
          enum: ["marketing", "hr", "finance", "product", "sales", "general"],
          description: "Business domain for pre-loaded archetypes and rules.",
        },
        rounds: {
          type: "number",
          description:
            "Number of simulation rounds (5-30). More rounds = deeper simulation but slower. Default: 15.",
          minimum: 5,
          maximum: 30,
        },
        context: {
          type: "string",
          description:
            "Optional additional context: company size, industry, current market position, etc.",
        },
      },
      required: ["question", "domain"],
    },
  },
  {
    name: "fracture_quick_pulse",
    description:
      "Fast 5-second market tension check. Analyzes a situation with 4 key archetypes and returns a tension score (0-100) plus the top 2 pressure points. Use this before committing to a decision.",
    inputSchema: {
      type: "object",
      properties: {
        situation: {
          type: "string",
          description:
            "The situation or decision to evaluate. Example: 'We are about to raise prices by 20%.'",
        },
        domain: {
          type: "string",
          enum: ["marketing", "hr", "finance", "product", "sales", "general"],
        },
      },
      required: ["situation", "domain"],
    },
  },
  {
    name: "fracture_list_rules",
    description:
      "List the market rules (with stability weights) for a given domain. Rules are the fundamental assumptions that govern how a market works. Low stability = ripe for disruption.",
    inputSchema: {
      type: "object",
      properties: {
        domain: {
          type: "string",
          enum: ["marketing", "hr", "finance", "product", "sales", "general"],
        },
      },
      required: ["domain"],
    },
  },
  {
    name: "fracture_archetypes",
    description:
      "List all 12 FRACTURE agent archetypes with their personalities, goals, and power levels. 8 Conformists defend existing rules; 4 Disruptors try to break them.",
    inputSchema: {
      type: "object",
      properties: {
        type: {
          type: "string",
          enum: ["all", "conformists", "disruptors"],
          description: "Filter by archetype type. Default: all.",
        },
      },
    },
  },
  {
    name: "fracture_history",
    description:
      "Retrieve past simulation results. Useful for comparing predictions over time or reviewing what the simulation said before a decision was made.",
    inputSchema: {
      type: "object",
      properties: {
        limit: {
          type: "number",
          description: "Number of past simulations to return. Default: 5.",
          minimum: 1,
          maximum: 50,
        },
        domain: {
          type: "string",
          enum: ["marketing", "hr", "finance", "product", "sales", "general"],
          description: "Filter by domain. Optional.",
        },
      },
    },
  },
  {
    name: "fracture_feedback",
    description:
      "Submit the real-world outcome of a past simulation. FRACTURE uses this to calibrate archetype accuracy over time — the more feedback you give, the more precise future simulations become.",
    inputSchema: {
      type: "object",
      properties: {
        simulation_id: {
          type: "string",
          description: "ID of the simulation to provide feedback for.",
        },
        outcome: {
          type: "string",
          enum: ["accurate", "partially_accurate", "inaccurate"],
          description: "How accurate was the simulation compared to reality?",
        },
        notes: {
          type: "string",
          description:
            "Optional notes on what actually happened vs what was predicted.",
        },
      },
      required: ["simulation_id", "outcome"],
    },
  },
];

// ─── Server setup ─────────────────────────────────────────────────────────────
const server = new Server(
  { name: "fracture", version: "1.0.0" },
  { capabilities: { tools: {} } }
);

server.setRequestHandler(ListToolsRequestSchema, async () => ({ tools: TOOLS }));

server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const { name, arguments: args } = request.params;

  try {
    let result: unknown;

    switch (name) {
      case "fracture_simulate":
        result = await callFracture("/simulate", args);
        break;

      case "fracture_quick_pulse":
        result = await callFracture("/pulse", args);
        break;

      case "fracture_list_rules":
        result = await callFracture(`/rules/${(args as { domain: string }).domain}`);
        break;

      case "fracture_archetypes":
        result = await callFracture(
          `/archetypes${(args as { type?: string }).type ? `?type=${(args as { type?: string }).type}` : ""}`
        );
        break;

      case "fracture_history":
        result = await callFracture(
          `/simulations?limit=${(args as { limit?: number }).limit ?? 5}${
            (args as { domain?: string }).domain
              ? `&domain=${(args as { domain?: string }).domain}`
              : ""
          }`
        );
        break;

      case "fracture_feedback":
        result = await callFracture("/feedback", args);
        break;

      default:
        throw new Error(`Unknown tool: ${name}`);
    }

    return {
      content: [
        {
          type: "text",
          text: JSON.stringify(result, null, 2),
        },
      ],
    };
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);

    // If FRACTURE is not running, give a helpful message
    if (message.includes("ECONNREFUSED") || message.includes("fetch failed")) {
      return {
        content: [
          {
            type: "text",
            text: `⚠️ FRACTURE is not running.\n\nStart it first:\n  • Windows: double-click fracture-windows-amd64.exe\n  • Linux/Mac: ./fracture\n\nThen try again. FRACTURE runs at http://localhost:3000`,
          },
        ],
        isError: true,
      };
    }

    return {
      content: [{ type: "text", text: `Error: ${message}` }],
      isError: true,
    };
  }
});

// ─── Start ────────────────────────────────────────────────────────────────────
async function main() {
  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error("FRACTURE MCP Server running — waiting for connections");
}

main().catch((err) => {
  console.error("Fatal:", err);
  process.exit(1);
});

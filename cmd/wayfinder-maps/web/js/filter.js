// Ticket classification shared by the toolbar, the card list and the map:
// who works a ticket (HITL or AFK), which actionability bucket it falls in,
// and whether it passes the current filters.
import {S} from "./state.js";

// mode follows the wayfinder skill: grilling and prototype need the human in
// the loop, research runs AFK, and a task can go either way.
export function mode(n) {
  if (n.type === "research") return "AFK";
  if (n.type === "task") return "either";
  return "HITL";
}

// ACTS are the actionability buckets, in list order. takeable is the frontier;
// needyou and agentready split it by who can pick the ticket up (a task sits
// in both). test decides membership.
export const ACTS = [
  {key: "takeable",   label: "takeable",    sq: "frontier", test: function(n) { return n.status === "frontier"; }},
  {key: "needyou",    label: "need you",    sq: "",         test: function(n) { return n.status === "frontier" && mode(n) !== "AFK"; }},
  {key: "agentready", label: "agent-ready", sq: "",         test: function(n) { return n.status === "frontier" && mode(n) !== "HITL"; }},
  {key: "claimed",    label: "in progress", sq: "claimed",  test: function(n) { return n.status === "claimed"; }},
  {key: "blocked",    label: "blocked",     sq: "blocked",  test: function(n) { return n.status === "blocked"; }},
  {key: "resolved",   label: "resolved",    sq: "resolved", test: function(n) { return n.status === "resolved"; }},
  {key: "oos",        label: "out of scope", sq: "oos",     test: function(n) { return n.status === "out_of_scope"; }}
];

export function actTest(key) {
  for (var i = 0; i < ACTS.length; i++) if (ACTS[i].key === key) return ACTS[i].test;
  return function() { return true; };
}

// TYPES is the display order for type chips.
export const TYPES = ["grilling", "research", "prototype", "task"];

export function filtering() { return !!(S.filter.type || S.filter.act || S.filter.q); }

export function matches(n) {
  var f = S.filter;
  if (f.type && n.type !== f.type) return false;
  if (f.act && !actTest(f.act)(n)) return false;
  if (f.q) {
    var q = f.q.toLowerCase().replace(/^#/, "");
    if (String(n.num) !== q && (n.title + "\n" + (n.body || "")).toLowerCase().indexOf(q) < 0) return false;
  }
  return true;
}

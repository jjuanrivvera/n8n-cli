#!/usr/bin/env python3
# Minimal mock n8n public API for the demo GIF — serves fake demo workflows so the
# recording shows clean, reproducible output with no real instance or secrets.
import json, sys
from http.server import BaseHTTPRequestHandler, HTTPServer

WF = [
  {"id":"a1","name":"crm-sync","active":True,"isArchived":False,"triggerCount":1,"updatedAt":"2026-06-22T10:00:00.000Z",
   "nodes":[{"id":"w","name":"Webhook","type":"n8n-nodes-base.webhook","webhookId":"crm","parameters":{"path":"crm"}},
            {"id":"h","name":"HTTP","type":"n8n-nodes-base.httpRequest","parameters":{"url":"https://api.crm.example/contacts"}}],
   "connections":{"Webhook":{"main":[[{"node":"HTTP","type":"main","index":0}]]}},"settings":{"executionOrder":"v1"}},
  {"id":"a2","name":"alert-router","active":True,"isArchived":False,"triggerCount":1,"updatedAt":"2026-06-22T11:30:00.000Z",
   "nodes":[{"id":"s","name":"Schedule","type":"n8n-nodes-base.scheduleTrigger","parameters":{}},
            {"id":"slack","name":"Slack","type":"n8n-nodes-base.slack","parameters":{}}],
   "connections":{"Schedule":{"main":[[{"node":"Slack","type":"main","index":0}]]}},"settings":{"executionOrder":"v1"}},
]
BYID={w["id"]:w for w in WF}

class H(BaseHTTPRequestHandler):
    def log_message(self,*a): pass
    def _send(self,obj,code=200):
        b=json.dumps(obj).encode(); self.send_response(code)
        self.send_header("Content-Type","application/json"); self.end_headers(); self.wfile.write(b)
    def do_GET(self):
        p=self.path.split("?")[0]
        if p=="/api/v1/workflows": return self._send({"data":WF,"nextCursor":None})
        if p.startswith("/api/v1/workflows/"): return self._send(BYID.get(p.rsplit("/",1)[1],WF[0]))
        if p in ("/api/v1/tags","/api/v1/variables","/api/v1/credentials"): return self._send({"data":[],"nextCursor":None})
        return self._send({"data":[],"nextCursor":None})
    def do_POST(self): self._send({"id":"new","name":"created"})
HTTPServer(("127.0.0.1",8799),H).serve_forever()

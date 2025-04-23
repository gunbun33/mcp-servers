from fastapi import FastAPI, Request, HTTPException, Depends, status
from fastapi.responses import StreamingResponse, JSONResponse
from fastapi.middleware.cors import CORSMiddleware
from fastapi.middleware.gzip import GZipMiddleware
import asyncio
import json
import os
import sys
import time
from typing import Dict, List, Any, Optional, Union
from pydantic import BaseModel, Field
import uuid
from loguru import logger
from prometheus_client import Counter, Histogram, start_http_server
import traceback
from datetime import datetime
import jsonschema
from contextlib import asynccontextmanager

# Configure logging
LOG_LEVEL = os.getenv("LOG_LEVEL", "INFO").upper()
logger.remove()
logger.add(sys.stderr, level=LOG_LEVEL)
logger.add("mcp_server.log", rotation="10 MB", level=LOG_LEVEL)

# Metrics
REQUEST_COUNT = Counter('mcp_request_count', 'Count of requests received', ['method', 'endpoint'])
REQUEST_LATENCY = Histogram('mcp_request_latency_seconds', 'Request latency in seconds', ['method', 'endpoint'])

# Configuration
PORT = int(os.getenv("PORT", "8080"))
HOST = os.getenv("HOST", "0.0.0.0")
DEBUG = os.getenv("DEBUG", "False").lower() in ("true", "1", "t")
METRICS_PORT = int(os.getenv("METRICS_PORT", "8081"))
ALLOWED_ORIGINS = os.getenv("ALLOWED_ORIGINS", "*").split(",")

# MCP Protocol settings
MCP_SERVER_NAME = os.getenv("MCP_SERVER_NAME", "Python FastAPI MCP")
MCP_SERVER_VERSION = os.getenv("MCP_SERVER_VERSION", "1.0.0")
MCP_PROTOCOL_VERSION = "2.0"

# Start metrics server
try:
    start_http_server(METRICS_PORT)
    logger.info(f"Metrics server started on port {METRICS_PORT}")
except Exception as e:
    logger.error(f"Failed to start metrics server: {e}")

# Pydantic models for request validation
class MCPRequest(BaseModel):
    jsonrpc: str = "2.0"
    id: Union[int, str]
    method: str
    params: Optional[Dict[str, Any]] = None

class MCPResponse(BaseModel):
    jsonrpc: str = "2.0"
    id: Union[int, str]
    result: Optional[Dict[str, Any]] = None
    error: Optional[Dict[str, Any]] = None

# Lifespan events
@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    logger.info("MCP Server starting up")
    # Add any startup code here (database connections, etc.)
    yield
    # Shutdown
    logger.info("MCP Server shutting down")
    # Add any cleanup code here

# Create FastAPI app
app = FastAPI(
    title="MCP Python Agent",
    description="Model Context Protocol server for VS Code",
    version="1.0.0",
    lifespan=lifespan,
    docs_url="/docs" if DEBUG else None,
    redoc_url="/redoc" if DEBUG else None,
)

# Add middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=ALLOWED_ORIGINS,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)
app.add_middleware(GZipMiddleware, minimum_size=1000)

# Request timing middleware
@app.middleware("http")
async def add_timing_and_logging(request: Request, call_next):
    start_time = time.time()
    method = request.method
    path = request.url.path
    
    # Log the request
    request_id = str(uuid.uuid4())
    logger.info(f"Request {request_id}: {method} {path}")
    
    try:
        response = await call_next(request)
        
        # Update metrics
        REQUEST_COUNT.labels(method=method, endpoint=path).inc()
        REQUEST_LATENCY.labels(method=method, endpoint=path).observe(time.time() - start_time)
        
        # Log the response
        logger.info(f"Response {request_id}: {response.status_code} - {time.time() - start_time:.4f}s")
        return response
    except Exception as e:
        logger.error(f"Error {request_id}: {str(e)}\n{traceback.format_exc()}")
        return JSONResponse(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            content={
                "jsonrpc": "2.0",
                "id": None,
                "error": {
                    "code": -32603,
                    "message": "Internal server error",
                    "data": {"detail": str(e)} if DEBUG else None
                }
            }
        )

@app.get("/health")
async def health_check():
    """Health check endpoint for monitoring and container orchestration.
    
    Returns a 200 OK response with server status information.
    """
    try:
        # You can add more health checks here (database connectivity, etc.)
        return {
            "status": "ok",
            "timestamp": datetime.utcnow().isoformat(),
            "version": MCP_SERVER_VERSION,
            "service": "mcp-python-agent"
        }
    except Exception as e:
        logger.error(f"Health check failed: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Health check failed"
        )

@app.get("/sse")
async def sse(request: Request):
    """Server-Sent Events endpoint for VS Code extension.
    
    This endpoint establishes a long-lived connection with the VS Code extension
    to push server events, notifications, and heartbeats.
    """
    # Log connection
    client_id = str(uuid.uuid4())
    client_info = request.headers.get("User-Agent", "Unknown client")
    logger.info(f"SSE connection established: {client_id} - {client_info}")
    
    # Check for VS Code specific headers
    is_vscode = "vscode" in client_info.lower()
    
    async def event_stream():
        try:
            # Send ready event
            yield 'data: {"type": "ready", "clientId": "' + client_id + '"}\n\n'
            await asyncio.sleep(1)
            
            # Send capabilities
            response = {
                "jsonrpc": "2.0",
                "id": 1,
                "result": {
                    "capabilities": {
                        "serverName": MCP_SERVER_NAME,
                        "serverVersion": MCP_SERVER_VERSION,
                        "tools": [
                            {
                                "name": "list_tables",
                                "description": "List all available tables",
                                "parameters": {
                                    "type": "object",
                                    "properties": {}
                                }
                            },
                            {
                                "name": "discover_data",
                                "description": "Discover data in tables",
                                "parameters": {
                                    "type": "object",
                                    "properties": {
                                        "table": {
                                            "type": "string",
                                            "description": "Table name to discover"
                                        }
                                    },
                                    "required": ["table"]
                                }
                            },
                            {
                                "name": "prepare_query",
                                "description": "Prepare a SQL query",
                                "parameters": {
                                    "type": "object",
                                    "properties": {
                                        "query": {
                                            "type": "string",
                                            "description": "SQL query to prepare"
                                        }
                                    },
                                    "required": ["query"]
                                }
                            },
                            {
                                "name": "query",
                                "description": "Execute a SQL query",
                                "parameters": {
                                    "type": "object",
                                    "properties": {
                                        "query": {
                                            "type": "string",
                                            "description": "SQL query to execute"
                                        }
                                    },
                                    "required": ["query"]
                                }
                            }
                        ],
                        "capabilities": {
                            "supportedLanguages": ["sql", "python"],
                            "supportsNotebooks": true,
                            "supportsInlineCompletions": true
                        }
                    }
                }
            }
            yield f'data: {json.dumps(response)}\n\n'
            
            # Heartbeat loop with error handling
            heartbeat_interval = 10  # seconds
            missed_heartbeats = 0
            max_missed_heartbeats = 3
            
            while True:
                try:
                    await asyncio.sleep(heartbeat_interval)
                    timestamp = datetime.utcnow().isoformat()
                    yield f'data: {{"type": "heartbeat", "timestamp": "{timestamp}", "clientId": "{client_id}"}}\n\n'
                    missed_heartbeats = 0
                except asyncio.CancelledError:
                    logger.info(f"SSE connection closed for client {client_id}")
                    break
                except Exception as e:
                    missed_heartbeats += 1
                    logger.error(f"Error sending heartbeat to client {client_id}: {e}")
                    if missed_heartbeats >= max_missed_heartbeats:
                        logger.warning(f"Too many missed heartbeats for client {client_id}, closing connection")
                        break
        except Exception as e:
            logger.error(f"SSE stream error for client {client_id}: {e}\n{traceback.format_exc()}")
            # Send error event before closing
            yield f'data: {{"type": "error", "message": "Stream error", "detail": "{str(e)}"}}\n\n'
        finally:
            logger.info(f"SSE connection terminated: {client_id}")
    
    return StreamingResponse(
        event_stream(), 
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "X-Accel-Buffering": "no",  # Disable proxy buffering
        }
    )

@app.post("/")
async def mcp_post(request: Request):
    """Main MCP endpoint that handles JSON-RPC requests from VS Code.
    
    This endpoint implements the Model Context Protocol for VS Code extensions.
    It validates the incoming requests, processes them, and returns appropriate responses.
    """
    request_id = str(uuid.uuid4())
    start_time = time.time()
    
    try:
        # Parse and validate the request body
        try:
            body = await request.json()
            # Validate against JSON-RPC schema
            mcp_request = MCPRequest(**body)
            method = mcp_request.method
            request_id_from_client = mcp_request.id
            params = mcp_request.params or {}
            
            logger.info(f"MCP request {request_id}: method={method}, id={request_id_from_client}")
        except Exception as e:
            logger.error(f"Invalid request {request_id}: {e}")
            return JSONResponse(
                status_code=status.HTTP_400_BAD_REQUEST,
                content={
                    "jsonrpc": "2.0",
                    "id": body.get("id") if isinstance(body, dict) else None,
                    "error": {
                        "code": -32700,
                        "message": "Parse error",
                        "data": {"detail": str(e)} if DEBUG else None
                    }
                }
            )
        
        # Process the request based on the method
        if method == "initialize":
            # VS Code extension initialization
            client_info = params.get("clientInfo", {})
            client_name = client_info.get("name", "Unknown")
            client_version = client_info.get("version", "Unknown")
            
            logger.info(f"Client initialized: {client_name} v{client_version}")
            
            return JSONResponse({
                "jsonrpc": "2.0",
                "id": request_id_from_client,
                "result": {
                    "capabilities": {
                        "serverName": MCP_SERVER_NAME,
                        "serverVersion": MCP_SERVER_VERSION,
                        "tools": [
                            {
                                "name": "list_tables",
                                "description": "List all available tables",
                                "parameters": {
                                    "type": "object",
                                    "properties": {}
                                }
                            },
                            {
                                "name": "discover_data",
                                "description": "Discover data in tables",
                                "parameters": {
                                    "type": "object",
                                    "properties": {
                                        "table": {
                                            "type": "string",
                                            "description": "Table name to discover"
                                        }
                                    },
                                    "required": ["table"]
                                }
                            },
                            {
                                "name": "prepare_query",
                                "description": "Prepare a SQL query",
                                "parameters": {
                                    "type": "object",
                                    "properties": {
                                        "query": {
                                            "type": "string",
                                            "description": "SQL query to prepare"
                                        }
                                    },
                                    "required": ["query"]
                                }
                            },
                            {
                                "name": "query",
                                "description": "Execute a SQL query",
                                "parameters": {
                                    "type": "object",
                                    "properties": {
                                        "query": {
                                            "type": "string",
                                            "description": "SQL query to execute"
                                        }
                                    },
                                    "required": ["query"]
                                }
                            }
                        ],
                        "capabilities": {
                            "supportedLanguages": ["sql", "python"],
                            "supportsNotebooks": true,
                            "supportsInlineCompletions": true
                        }
                    }
                }
            })
        elif method == "shutdown":
            # VS Code extension shutdown
            logger.info(f"Client requested shutdown")
            return JSONResponse({
                "jsonrpc": "2.0",
                "id": request_id_from_client,
                "result": None
            })
        elif method == "list_tables":
            # Example response with proper error handling
            try:
                # Here you would implement actual table listing logic
                tables = ["users", "products", "orders"]
                
                return JSONResponse({
                    "jsonrpc": "2.0",
                    "id": request_id_from_client,
                    "result": {
                        "tables": tables
                    }
                })
            except Exception as e:
                logger.error(f"Error listing tables: {e}\n{traceback.format_exc()}")
                return JSONResponse({
                    "jsonrpc": "2.0",
                    "id": request_id_from_client,
                    "error": {
                        "code": -32000,
                        "message": "Internal error",
                        "data": {"detail": str(e)} if DEBUG else None
                    }
                })
        elif method == "discover_data":
            try:
                # Validate required parameters
                table = params.get("table")
                if not table:
                    return JSONResponse({
                        "jsonrpc": "2.0",
                        "id": request_id_from_client,
                        "error": {
                            "code": -32602,
                            "message": "Invalid params",
                            "data": {"detail": "Missing required parameter: table"}
                        }
                    })
                
                # Here you would implement actual table discovery logic
                # Example response
                return JSONResponse({
                    "jsonrpc": "2.0",
                    "id": request_id_from_client,
                    "result": {
                        "columns": [
                            {"name": "id", "type": "integer"},
                            {"name": "name", "type": "string"},
                            {"name": "created_at", "type": "timestamp"}
                        ]
                    }
                })
            except Exception as e:
                logger.error(f"Error discovering data: {e}\n{traceback.format_exc()}")
                return JSONResponse({
                    "jsonrpc": "2.0",
                    "id": request_id_from_client,
                    "error": {
                        "code": -32000,
                        "message": "Internal error",
                        "data": {"detail": str(e)} if DEBUG else None
                    }
                })
        elif method == "prepare_query":
            try:
                # Validate required parameters
                query = params.get("query")
                if not query:
                    return JSONResponse({
                        "jsonrpc": "2.0",
                        "id": request_id_from_client,
                        "error": {
                            "code": -32602,
                            "message": "Invalid params",
                            "data": {"detail": "Missing required parameter: query"}
                        }
                    })
                
                # Here you would implement actual query preparation logic
                # Example response
                return JSONResponse({
                    "jsonrpc": "2.0",
                    "id": request_id_from_client,
                    "result": {
                        "prepared": True,
                        "parameters": []
                    }
                })
            except Exception as e:
                logger.error(f"Error preparing query: {e}\n{traceback.format_exc()}")
                return JSONResponse({
                    "jsonrpc": "2.0",
                    "id": request_id_from_client,
                    "error": {
                        "code": -32000,
                        "message": "Internal error",
                        "data": {"detail": str(e)} if DEBUG else None
                    }
                })
        elif method == "query":
            try:
                # Validate required parameters
                query = params.get("query")
                if not query:
                    return JSONResponse({
                        "jsonrpc": "2.0",
                        "id": request_id_from_client,
                        "error": {
                            "code": -32602,
                            "message": "Invalid params",
                            "data": {"detail": "Missing required parameter: query"}
                        }
                    })
                
                # Here you would implement actual query execution logic
                # Example response
                return JSONResponse({
                    "jsonrpc": "2.0",
                    "id": request_id_from_client,
                    "result": {
                        "columns": ["id", "name"],
                        "rows": [
                            [1, "Example"],
                            [2, "Test"]
                        ]
                    }
                })
            except Exception as e:
                logger.error(f"Error executing query: {e}\n{traceback.format_exc()}")
                return JSONResponse({
                    "jsonrpc": "2.0",
                    "id": request_id_from_client,
                    "error": {
                        "code": -32000,
                        "message": "Internal error",
                        "data": {"detail": str(e)} if DEBUG else None
                    }
                })
        else:
            # Method not found
            logger.warning(f"Method not found: {method}")
            return JSONResponse({
                "jsonrpc": "2.0",
                "id": request_id_from_client,
                "error": {
                    "code": -32601,
                    "message": "Method not found",
                    "data": {"method": method}
                }
            })
    except Exception as e:
        # Catch-all for unexpected errors
        logger.error(f"Unexpected error: {e}\n{traceback.format_exc()}")
        return JSONResponse(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            content={
                "jsonrpc": "2.0",
                "id": None,
                "error": {
                    "code": -32603,
                    "message": "Internal error",
                    "data": {"detail": str(e)} if DEBUG else None
                }
            }
        )
    finally:
        # Log request processing time
        logger.info(f"MCP request {request_id} processed in {time.time() - start_time:.4f}s")
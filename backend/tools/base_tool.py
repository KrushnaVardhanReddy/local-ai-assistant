"""
BaseTool — Abstract base for all BarnOwl AI agent tools.

To add a new tool:
  1. Create a new file in backend/tools/
  2. Define a class that inherits from BaseTool
  3. Implement: name, description, parameters, execute()
  4. The ToolRegistry will auto-discover it. Nothing else to change.
"""
from abc import ABC, abstractmethod
from typing import Any

class BaseTool(ABC):
    @property
    @abstractmethod
    def name(self) -> str:
        """Unique tool name — used as the function name in OpenAI schema."""
        ...

    @property
    @abstractmethod
    def description(self) -> str:
        """Human-readable description. The LLM reads this to decide when to use the tool."""
        ...

    @property
    @abstractmethod
    def parameters(self) -> dict:
        """JSON Schema for the tool's arguments. Must be valid OpenAI tool parameters spec."""
        ...

    @abstractmethod
    async def execute(self, **kwargs: Any) -> str:
        """
        Execute the tool with the given arguments.
        Must return a string result that will be fed back to the LLM.
        Never raise — catch exceptions and return an error string.
        """
        ...

    def to_openai_schema(self) -> dict:
        """Convert this tool to OpenAI function-calling JSON format."""
        return {
            "type": "function",
            "function": {
                "name": self.name,
                "description": self.description,
                "parameters": self.parameters,
            }
        }

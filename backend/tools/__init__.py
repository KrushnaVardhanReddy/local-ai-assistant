"""
ToolRegistry — Auto-discovers and registers all BaseTool subclasses
in the backend/tools/ directory. No manual registration needed.
"""
import importlib
import inspect
import pkgutil
from pathlib import Path
from .base_tool import BaseTool

class ToolRegistry:
    def __init__(self):
        self._tools: dict[str, BaseTool] = {}
        self._discover()

    def _discover(self):
        """Scan backend/tools/ for BaseTool subclasses and register them."""
        tools_dir = Path(__file__).parent
        for module_info in pkgutil.iter_modules([str(tools_dir)]):
            if module_info.name in ("base_tool", "__init__"):
                continue
            try:
                module = importlib.import_module(f"tools.{module_info.name}")
                for _, cls in inspect.getmembers(module, inspect.isclass):
                    if issubclass(cls, BaseTool) and cls is not BaseTool:
                        instance = cls()
                        self._tools[instance.name] = instance
            except Exception as e:
                print(f"[ToolRegistry] Failed to load tools/{module_info.name}.py: {e}")

    def get_all(self) -> list[BaseTool]:
        return list(self._tools.values())

    def get_schemas(self) -> list[dict]:
        return [t.to_openai_schema() for t in self._tools.values()]

    def get(self, name: str) -> BaseTool | None:
        return self._tools.get(name)

# Singleton
registry = ToolRegistry()

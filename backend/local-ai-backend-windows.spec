# -*- mode: python ; coding: utf-8 -*-
#
# Windows build spec — bundles CUDA 12 + cuDNN DLLs so that
# faster-whisper can use GPU acceleration on any NVIDIA machine
# without requiring the user to install CUDA themselves.
#
# Build:  pyinstaller local-ai-backend-windows.spec
# Output: dist/local-ai-backend-windows/   (folder, NOT a single file)
#         The folder contains the .exe + all bundled DLLs.

import os
import sys
from pathlib import Path

# ---------------------------------------------------------------------------
# Collect CUDA 12 DLLs from the CUDA Toolkit installed on the build machine.
# On GitHub Actions we install CUDA via the cuda-installer step; the default
# install path is C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v12.x\bin
# ---------------------------------------------------------------------------

CUDA_DLL_NAMES = [
    # cuBLAS — required by CTranslate2 / faster-whisper for GPU inference
    "cublas64_12.dll",
    "cublasLt64_12.dll",
    # cuDNN — required for whisper model operations on GPU
    "cudnn64_9.dll",
    "cudnn_ops64_9.dll",
    "cudnn_cnn64_9.dll",
    # CUDA runtime
    "cudart64_12.dll",
    # zlib (sometimes needed by cuDNN on Windows)
    "zlibwapi.dll",
]

cuda_binaries = []

# Search common CUDA install locations on Windows.
# CUDA_PATH is exported by the Jimver/cuda-toolkit GitHub Action and is the
# most reliable source — checked first.
cuda_search_dirs = []

cuda_env = os.environ.get("CUDA_PATH", "")
if cuda_env:
    cuda_search_dirs.append(os.path.join(cuda_env, "bin"))

cuda_search_dirs += [
    r"C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v12.6\bin",
    r"C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v12.5\bin",
    r"C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v12.4\bin",
    r"C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v12.3\bin",
    r"C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v12.2\bin",
    r"C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v12.1\bin",
    r"C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v12.0\bin",
    r"C:\Program Files\NVIDIA\CUDNN\v9\bin\12.6",
    r"C:\Program Files\NVIDIA\CUDNN\v9\bin",
]

# Also scan any PATH entry that looks CUDA/NVIDIA related
for p in os.environ.get("PATH", "").split(os.pathsep):
    if "cuda" in p.lower() or "nvidia" in p.lower():
        cuda_search_dirs.append(p)

# Resolve the DLLs
for dll_name in CUDA_DLL_NAMES:
    for search_dir in cuda_search_dirs:
        candidate = Path(search_dir) / dll_name
        if candidate.exists():
            cuda_binaries.append((str(candidate), "."))
            print(f"[spec] Bundling CUDA DLL: {candidate}")
            break
    else:
        print(f"[spec] WARNING: {dll_name} not found — GPU may not work without CUDA installed.")

print(f"[spec] Total CUDA DLLs bundled: {len(cuda_binaries)}")

# ---------------------------------------------------------------------------
# Main PyInstaller analysis
# ---------------------------------------------------------------------------

a = Analysis(
    ['app.py'],
    pathex=[],
    binaries=cuda_binaries,
    datas=[],
    hiddenimports=[
        'fastapi',
        'uvicorn',
        'uvicorn.logging',
        'uvicorn.loops',
        'uvicorn.loops.auto',
        'uvicorn.protocols',
        'uvicorn.protocols.http',
        'uvicorn.protocols.http.auto',
        'uvicorn.protocols.websockets',
        'uvicorn.protocols.websockets.auto',
        'uvicorn.lifespan',
        'uvicorn.lifespan.on',
        'faster_whisper',
        'ctranslate2',
        'websockets',
        'sounddevice',
        'chromadb',
        'sentence_transformers',
    ],
    hookspath=[],
    hooksconfig={},
    runtime_hooks=[],
    excludes=[
        'tkinter',
        'matplotlib',
        'PyQt5',
        'PyQt6',
        'IPython',
        'notebook',
        'pytest',
        'tensorboard',
    ],
    noarchive=False,
    optimize=0,
)

pyz = PYZ(a.pure)

exe = EXE(
    pyz,
    a.scripts,
    [],
    exclude_binaries=True,
    name='local-ai-backend-windows',
    debug=False,
    bootloader_ignore_signals=False,
    strip=False,
    upx=False,
    upx_exclude=[],
    runtime_tmpdir=None,
    console=True,
    disable_windowed_traceback=False,
    argv_emulation=False,
    target_arch=None,
    codesign_identity=None,
    entitlements_file=None,
)

coll = COLLECT(
    exe,
    a.binaries,
    a.datas,
    strip=False,
    upx=False,
    upx_exclude=[],
    name='local-ai-backend-windows',
)

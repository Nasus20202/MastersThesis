"""Logging for every entrypoint: one format on stderr, and a run's own log
file kept next to its results."""

from __future__ import annotations

import contextlib
import logging
from pathlib import Path

FORMAT = "%(asctime)s %(levelname)s %(name)s: %(message)s"


def setup() -> None:
    """Called once by each entrypoint, before anything logs."""
    logging.basicConfig(format=FORMAT, level=logging.INFO, force=True)
    # Progress bars write one line per update and would bury the log.
    with contextlib.suppress(ImportError):
        from huggingface_hub.utils import disable_progress_bars

        disable_progress_bars()
    with contextlib.suppress(ImportError):
        from transformers.utils import logging as transformers_logging

        transformers_logging.disable_progress_bar()


class LogFile:
    """Also writes everything logged while open (``with``) to ``path``."""

    def __init__(self, path: Path):
        self.path = path
        self.handler: logging.FileHandler | None = None
        self.level = logging.NOTSET

    def __enter__(self) -> LogFile:
        self.path.parent.mkdir(parents=True, exist_ok=True)
        self.handler = logging.FileHandler(self.path)
        self.handler.setFormatter(logging.Formatter(FORMAT))
        root = logging.getLogger()
        self.level = root.level
        root.setLevel(min(self.level, logging.INFO))  # the file always gets the progress lines
        root.addHandler(self.handler)
        return self

    def __exit__(self, *exc: object) -> None:
        root = logging.getLogger()
        if self.handler:
            root.removeHandler(self.handler)
            self.handler.close()
        root.setLevel(self.level)

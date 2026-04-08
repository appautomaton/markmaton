from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any, Dict, List, Mapping, Optional


@dataclass
class ConvertOptions:
    only_main_content: bool = True
    include_selectors: List[str] = field(default_factory=list)
    exclude_selectors: List[str] = field(default_factory=list)

    def to_dict(self) -> Dict[str, Any]:
        return {
            "only_main_content": self.only_main_content,
            "include_selectors": list(self.include_selectors),
            "exclude_selectors": list(self.exclude_selectors),
        }


@dataclass
class ConvertRequest:
    html: str
    url: Optional[str] = None
    final_url: Optional[str] = None
    content_type: Optional[str] = None
    options: ConvertOptions = field(default_factory=ConvertOptions)

    def to_payload(self) -> Dict[str, Any]:
        payload: Dict[str, Any] = {
            "html": self.html,
            "options": self.options.to_dict(),
        }
        if self.url:
            payload["url"] = self.url
        if self.final_url:
            payload["final_url"] = self.final_url
        if self.content_type:
            payload["content_type"] = self.content_type
        return payload


@dataclass
class Metadata:
    title: str = ""
    description: str = ""
    canonical_url: str = ""
    language: str = ""
    author: str = ""
    og_title: str = ""
    og_description: str = ""
    extras: Dict[str, str] = field(default_factory=dict)

    @classmethod
    def from_dict(cls, value: Mapping[str, Any]) -> "Metadata":
        return cls(
            title=str(value.get("title", "") or ""),
            description=str(value.get("description", "") or ""),
            canonical_url=str(value.get("canonical_url", "") or ""),
            language=str(value.get("language", "") or ""),
            author=str(value.get("author", "") or ""),
            og_title=str(value.get("og_title", "") or ""),
            og_description=str(value.get("og_description", "") or ""),
            extras=dict(value.get("extras", {}) or {}),
        )


@dataclass
class Quality:
    text_length: int = 0
    paragraph_count: int = 0
    link_count: int = 0
    image_count: int = 0
    title_present: bool = False
    link_density: float = 0.0
    quality_score: float = 0.0
    used_main_content: bool = True
    fallback_used: bool = False

    @classmethod
    def from_dict(cls, value: Mapping[str, Any]) -> "Quality":
        return cls(
            text_length=int(value.get("text_length", 0) or 0),
            paragraph_count=int(value.get("paragraph_count", 0) or 0),
            link_count=int(value.get("link_count", 0) or 0),
            image_count=int(value.get("image_count", 0) or 0),
            title_present=bool(value.get("title_present", False)),
            link_density=float(value.get("link_density", 0.0) or 0.0),
            quality_score=float(value.get("quality_score", 0.0) or 0.0),
            used_main_content=bool(value.get("used_main_content", True)),
            fallback_used=bool(value.get("fallback_used", False)),
        )


@dataclass
class ConvertResponse:
    markdown: str
    html_clean: str
    metadata: Metadata
    links: List[str]
    images: List[str]
    quality: Quality

    @classmethod
    def from_dict(cls, value: Mapping[str, Any]) -> "ConvertResponse":
        return cls(
            markdown=str(value.get("markdown", "") or ""),
            html_clean=str(value.get("html_clean", "") or ""),
            metadata=Metadata.from_dict(value.get("metadata", {}) or {}),
            links=list(value.get("links", []) or []),
            images=list(value.get("images", []) or []),
            quality=Quality.from_dict(value.get("quality", {}) or {}),
        )

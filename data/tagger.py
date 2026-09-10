import argparse
import json
import os
import shutil
import sys
import exiftool

# Default ExifTool executable path (Windows default, with system PATH fallback)
EXIFTOOL_PATH = r"C:\Program Files\ExifTool\exiftool.exe"
if not os.path.exists(EXIFTOOL_PATH):
    EXIFTOOL_PATH = shutil.which("exiftool") or EXIFTOOL_PATH


PLATFORM_PRESETS = {
    "unsplash": {
        "name": "Unsplash",
        "language": "en",
        "max_tags": None,  # All possible keywords
        "summary": "All possible keywords in English",
    },
    "pexels": {
        "name": "Pexels",
        "language": "en",
        "max_tags": 30,
        "summary": "Up to 30 keywords in English",
    },
    "pixabay": {
        "name": "Pixabay",
        "language": "es",
        "max_tags": 25,
        "summary": "Up to 25 keywords in Spanish",
    },
}


def parse_metadata_entry(item, language="en", max_tags=None):
    """
    Extract, deduplicate, and limit tags and description according to language and preset.
    Supports both new format (description_en/es, tags_en/es) and legacy format (tags, description).
    """
    # Select tags and description based on language
    if language == "es":
        raw_tags = item.get("tags_es") or item.get("tags") or item.get("tags_en") or ""
        description = (
            item.get("description_es")
            or item.get("description")
            or item.get("description_en")
            or ""
        )
    else:  # default "en"
        raw_tags = item.get("tags_en") or item.get("tags") or item.get("tags_es") or ""
        description = (
            item.get("description_en")
            or item.get("description")
            or item.get("description_es")
            or ""
        )

    # Parse tags to list
    if isinstance(raw_tags, str):
        raw_list = [t.strip() for t in raw_tags.split(",") if t.strip()]
    elif isinstance(raw_tags, (list, tuple)):
        raw_list = [str(t).strip() for t in raw_tags if str(t).strip()]
    else:
        raw_list = []

    # Case-insensitive deduplication while preserving order
    seen = set()
    tag_list = []
    for t in raw_list:
        t_low = t.lower()
        if t_low not in seen:
            seen.add(t_low)
            tag_list.append(t)

    # Apply tag limit if specified
    if max_tags is not None and max_tags > 0 and len(tag_list) > max_tags:
        tag_list = tag_list[:max_tags]

    desc_str = str(description).strip() if description else ""

    return tag_list, desc_str


def update_image_metadata(
    json_path,
    root_folder,
    platform="unsplash",
    language=None,
    max_tags=None,
    dry_run=False,
):
    stats = {
        "files_scanned": 0,
        "matches_found": 0,
        "updated": 0,
        "skipped_dry_run": 0,
        "errors": 0,
        "error_details": [],
    }

    # Resolve platform preset settings
    platform_key = platform.lower() if platform else "unsplash"
    preset = PLATFORM_PRESETS.get(platform_key, PLATFORM_PRESETS["unsplash"])

    selected_lang = language if language is not None else preset["language"]
    selected_max_tags = max_tags if max_tags is not None else preset["max_tags"]
    platform_name = preset["name"]

    try:
        with open(json_path, "r", encoding="utf-8") as f:
            tag_data = json.load(f)

        # Map basename to item data (case-insensitive)
        metadata_map = {}
        for item in tag_data:
            if "filename" in item:
                clean_name = os.path.basename(item["filename"]).lower()
                metadata_map[clean_name] = item

    except Exception as e:
        print(f"[ERROR] Failed to load JSON '{json_path}': {e}")
        return stats

    mode_label = "[DRY RUN]" if dry_run else "[LIVE]"
    limit_str = (
        f"{selected_max_tags} tags" if selected_max_tags else "All (Unlimited)"
    )

    print(f"--- Starting Metadata Update {mode_label} ---")
    print(f"Platform Preset  : {platform_name} ({preset['summary']})")
    print(f"Language         : {selected_lang.upper()}")
    print(f"Max Tags Allowed : {limit_str}")
    print(f"JSON Metadata    : {os.path.abspath(json_path)}")
    print(f"Root Directory   : {os.path.abspath(root_folder)}\n")

    with exiftool.ExifToolHelper(
        executable=EXIFTOOL_PATH,
        encoding="utf-8",
    ) as et:
        for current_root, _, files in os.walk(root_folder):
            for file in files:
                stats["files_scanned"] += 1
                lookup_key = file.lower()

                if lookup_key in metadata_map:
                    stats["matches_found"] += 1
                    file_path = os.path.normpath(
                        os.path.join(current_root, file)
                    )
                    item_data = metadata_map[lookup_key]

                    tag_list, description = parse_metadata_entry(
                        item_data,
                        language=selected_lang,
                        max_tags=selected_max_tags,
                    )

                    # Build metadata payload
                    tags_payload = {
                        "XMP-dc:Subject": tag_list,
                        "IPTC:Keywords": tag_list,
                        "EXIF:XPKeywords": ", ".join(tag_list),
                    }
                    if description:
                        tags_payload["XMP-dc:Description"] = description
                        tags_payload["IPTC:Caption-Abstract"] = description
                        tags_payload["EXIF:ImageDescription"] = description

                    desc_preview = (
                        f" | Desc: '{description[:50]}...'"
                        if description
                        else " | No description"
                    )

                    if dry_run:
                        print(
                            f"[DRY RUN] Would update: {file_path} ({len(tag_list)} tags{desc_preview})"
                        )
                        if description:
                            print(f"          Description: {description}")
                        print(f"          Tags: {tag_list}")
                        stats["skipped_dry_run"] += 1
                    else:
                        try:
                            # Parameters force UTF-8 filename & character set handling
                            et.set_tags(
                                file_path,
                                tags=tags_payload,
                                params=[
                                    "-overwrite_original",
                                    "-charset",
                                    "filename=utf8",
                                    "-codedcharacterset=utf8",
                                ],
                            )
                            print(
                                f"[UPDATED] ({len(tag_list)} tags{desc_preview}) -> {file_path}"
                            )
                            stats["updated"] += 1

                        except Exception as e:
                            stats["errors"] += 1
                            stats["error_details"].append((file_path, str(e)))
                            print(
                                f"[ERROR] Failed writing metadata for {file_path}: {e}"
                            )

    print("\n========================================")
    print(f"       EXECUTION SUMMARY {mode_label}")
    print("========================================")
    print(f"Platform            : {platform_name}")
    print(f"Language            : {selected_lang.upper()}")
    print(f"Total Files Scanned : {stats['files_scanned']}")
    print(f"Matching Files Found: {stats['matches_found']}")
    if dry_run:
        print(f"Projected Updates   : {stats['skipped_dry_run']}")
    else:
        print(f"Successfully Updated: {stats['updated']}")
    print(f"Errors              : {stats['errors']}")
    print("========================================\n")
    return stats


def build_parser():
    epilog_text = """
===============================================================================
PLATFORM PRESETS & RULES:
  unsplash : English language (tags_en, description_en).
             Includes ALL available keywords (unlimited).
  pexels   : English language (tags_en, description_en).
             Limits keywords to a maximum of 30.
  pixabay  : Spanish language (tags_es, description_es).
             Limits keywords to a maximum of 25.

METADATA FIELDS UPDATED:
  - Keywords / Tags:
      * XMP-dc:Subject        (Adobe Lightroom & modern DAM software)
      * IPTC:Keywords         (Standard IPTC stock agency keywords)
      * EXIF:XPKeywords       (Windows Explorer keywords)
  - Description / Caption:
      * XMP-dc:Description    (Adobe Lightroom Caption / Description)
      * IPTC:Caption-Abstract (Standard IPTC Caption)
      * EXIF:ImageDescription (Standard EXIF description)

USAGE EXAMPLES:
  1. Preview changes safely without touching files (dry run):
     python tagger.py -p unsplash --dry-run

  2. Apply metadata for Unsplash (English, unlimited keywords):
     python tagger.py --platform unsplash

  3. Apply metadata for Pexels (English, max 30 keywords):
     python tagger.py --platform pexels

  4. Apply metadata for Pixabay (Spanish, max 25 keywords):
     python tagger.py --platform pixabay

  5. Specify custom JSON file and target photo directory:
     python tagger.py -p pixabay -j new_tags.json -d "D:\\Photos"
===============================================================================
"""

    # Resolve default directory
    target_dir = os.getenv("DEFAULT_PHOTOS_DIR", ".")
    if not os.path.exists(target_dir):
        target_dir = "."

    # Resolve default JSON path
    json_file = "new_tags.json"
    if not os.path.exists(json_file):
        script_dir = os.path.dirname(os.path.abspath(__file__))
        cand_json = os.path.join(script_dir, json_file)
        if os.path.exists(cand_json):
            json_file = cand_json

    parser = argparse.ArgumentParser(
        description="Update image metadata (tags & descriptions) for Lightroom & Stock Platforms.",
        epilog=epilog_text,
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    parser.add_argument(
        "-p",
        "--platform",
        choices=["unsplash", "pexels", "pixabay"],
        default="unsplash",
        help="Stock platform preset to apply (default: %(default)s).\n"
        "  - unsplash: All keywords in English\n"
        "  - pexels:   30 keywords in English\n"
        "  - pixabay:  25 keywords in Spanish",
    )
    parser.add_argument(
        "-j",
        "--json",
        default=json_file,
        help="Path to JSON file with tags and descriptions (default: %(default)s)",
    )
    parser.add_argument(
        "-d",
        "--dir",
        default=target_dir,
        help="Root folder containing photos (searched recursively) (default: %(default)s)",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        default=False,
        help="Simulate execution without modifying any image files on disk",
    )
    return parser


if __name__ == "__main__":
    parser = build_parser()
    args = parser.parse_args()

    update_image_metadata(
        json_path=args.json,
        root_folder=args.dir,
        platform=args.platform,
        dry_run=args.dry_run,
    )
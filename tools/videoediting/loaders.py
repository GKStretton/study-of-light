import logging
from pathlib import Path
import typing
import os
from dataclasses import dataclass

from PIL import Image
import machinepb.machine as pb

import yaml


def get_session_metadata(base_dir: str, session_number: int):
    filename = "{}_session.yml".format(session_number)
    path = os.path.join(base_dir, "session_metadata", filename)
    yml = None
    with open(path, 'r') as f:
        yml = yaml.load(f, Loader=yaml.FullLoader)
    logging.info("Loaded session metadata: {}\n".format(yml))
    return yml


def get_session_content_path(base_dir: str, session_number: int):
    return os.path.join(base_dir, "session_content", str(session_number))


def get_state_reports(base_dir: str, session_number: int) -> typing.Tuple[float, pb.StateReport]:
    content_path = get_session_content_path(base_dir, session_number)
    raw_state_reports = None
    with open(os.path.join(content_path, "state-reports.yml"), 'r') as f:
        raw_state_reports = yaml.load(f, yaml.FullLoader)

    state_report_list = pb.StateReportList().from_dict(raw_state_reports).state_reports
    logging.info("Loaded {} state report entries\n".format(len(state_report_list)))

    state_reports: typing.Tuple[float, pb.StateReport] = []
    for _, v in enumerate(state_report_list):
        report = v
        report_ts = float(report.timestamp_unix_micros) / 1.0e6
        state_reports.append((report_ts, report))

    return state_reports


def get_content_plan(base_dir: str, session_number: int) -> pb.ContentTypeStatuses:
    content_path = get_session_content_path(base_dir, session_number)
    with open(os.path.join(content_path, "content_plan.yml"), 'r') as f:
        raw_plan = yaml.load(f, yaml.FullLoader)

    return pb.ContentTypeStatuses().from_dict(raw_plan)


def get_selected_dslr_image_path(base_dir: str, session_number: int, image_choice: str) -> str:
    filename = f"{image_choice}.jpg"
    return os.path.join(base_dir, "session_content", str(session_number), "dslr/post", filename)


def get_selected_dslr_image(base_dir: str, session_number: int, image_choice: str) -> Image.Image:
    return Image.open(get_selected_dslr_image_path(base_dir, session_number, image_choice))

# look up creationtime of selected.jpg dslr/post image


def get_selected_dslr_realpath(base_dir: str, session_number: int) -> str:
    linkname = f"selected.jpg"
    path = os.path.join(base_dir, "session_content", str(session_number), "dslr/post", linkname)

    # resolves symlink
    return os.path.realpath(path)


def get_selected_dslr_image_number(base_dir: str, session_number: int) -> int:
    real_path = get_selected_dslr_realpath(base_dir, session_number)
    name = Path(real_path).stem

    return int(name)


@dataclass
class MiscData:
    """Miscellaneous data to be passed to the pipeline"""
    selected_dslr_number: int = 0
    start_timestamp_s: int = 0


def get_misc_data(base_dir: str, session_number: int, state_reports: typing.Tuple[float, pb.StateReport]) -> MiscData:
    return MiscData(
        selected_dslr_number=get_selected_dslr_image_number(base_dir, session_number),
        start_timestamp_s=state_reports[0][0]
    )


def get_profile_snapshot(base_dir: str, session_number: int) -> pb.SystemVialConfigurationSnapshot:
    content_path = get_session_content_path(base_dir, session_number)
    with open(os.path.join(content_path, "vial-profiles.yml"), 'r') as f:
        raw_profiles = yaml.load(f, yaml.FullLoader)

    return pb.SystemVialConfigurationSnapshot().from_dict(raw_profiles)

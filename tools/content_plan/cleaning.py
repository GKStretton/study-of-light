from machinepb import machine as pb
from content_plan.loader import *


def buildCleaning(base_dir: str, n_str: str, unlisted: bool = False) -> pb.ContentTypeStatus:
    ct = pb.ContentType.CONTENT_TYPE_CLEANING
    raw_title, raw_description = get_random_title_and_description(ct)
    music_file, music_name = get_music(base_dir, ct)

    s = pb.ContentTypeStatus(
        raw_title=raw_title,
        raw_description=raw_description,
        caption=get_caption(ct),
        music_file=music_file,
        music_name=music_name,
    )

    platform = pb.SocialPlatform.SOCIAL_PLATFORM_YOUTUBE
    s.posts.append(pb.Post(
        platform=platform,
        title=append_title_hashtags(f"{n_str} (end): {s.raw_title}", ct, platform),
        description=f"{get_hashtags(ct, platform)}\n\n{s.raw_description}\n\n{get_common_text(ct, platform)}\n\n🎵 {music_name}",
        crosspost=False,
        scheduled_unix_timetamp=get_schedule_timestamp(ct),
        unlisted=unlisted,
    ))

    # platform = pb.SocialPlatform.SOCIAL_PLATFORM_TIKTOK
    # s.posts.append(pb.Post(
    #     platform=platform,
    #     title=append_title_hashtags(f"{s.raw_title} - {n_str}", ct, platform) + "\n\n" + get_common_text(ct, platform),
    #     description="N/A",
    #     crosspost=False,
    #     scheduled_unix_timetamp=get_schedule_timestamp(ct),
    #     unlisted=unlisted,
    # ))

    # platform = pb.SocialPlatform.SOCIAL_PLATFORM_INSTAGRAM
    # s.posts.append(pb.Post(
    #     platform=platform,
    #     title=f"{n_str} (end): {s.raw_title}\n\n{s.raw_description}\n\n{get_common_text(ct, platform)}\n\n{get_hashtags(ct, platform)}",
    #     description="N/A",
    #     crosspost=False,
    #     scheduled_unix_timetamp=get_schedule_timestamp(ct),
    #     unlisted=unlisted,
    # ))

    # platform = pb.SocialPlatform.SOCIAL_PLATFORM_FACEBOOK
    # s.posts.append(pb.Post(
    #     platform=platform,
    #     title=f"{s.raw_title} - {n_str}\n\n{s.raw_description}\n\n{get_common_text(ct, platform)}\n\n{get_hashtags(ct, platform)}",
    #     description="N/A",
    #     crosspost=False,
    #     scheduled_unix_timetamp=get_schedule_timestamp(ct),
    #     unlisted=unlisted,
    # ))

    # platform = pb.SocialPlatform.SOCIAL_PLATFORM_TWITTER
    # s.posts.append(pb.Post(
    #     platform=platform,
    #     title=f"{s.raw_title} - {n_str}\n\n{get_common_text(ct, platform)}\n\n{get_hashtags(ct, platform)}",
    #     description="N/A",
    #     crosspost=False,
    #     scheduled_unix_timetamp=get_schedule_timestamp(ct),
    #     unlisted=unlisted,
    # ))

    return s

from openapi.models import shared
from openapi import SDK
from openapi.models.operations import *

from .common_helpers import *


def test_open_enums_round_trip():
    record_test("open-enums-round-trip")

    s = SDK(server_url=HTTPBIN_URL)
    assert s is not None

    res = s.enums.enums_post_open_enum_unrecognized(
        request=shared.ThemeRequestOpaque(
            color="purple",
            icon="tick",
            hero_width=2160,
        )
    )
    assert res is not None

    theme = res.json_
    assert theme is not None
    assert theme.color == shared.ThemeColor("purple")
    assert theme.icon == shared.ThemeIcon.TICK
    assert theme.hero_width == shared.ThemeHeroWidth(2160)

    rt_res = s.enums.enums_post_open_enum_unrecognized(
        request=shared.ThemeRequestOpaque(
            color=theme.color,
            icon=theme.icon,
            hero_width=theme.hero_width,
        )
    )
    assert rt_res is not None

    rt_theme = rt_res.json_
    assert rt_theme is not None
    assert rt_theme.color == shared.ThemeColor("purple")
    assert rt_theme.icon == shared.ThemeIcon.TICK
    assert rt_theme.hero_width == shared.ThemeHeroWidth(2160)

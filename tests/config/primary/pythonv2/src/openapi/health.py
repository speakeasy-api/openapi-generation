# This incomplete file will trick the generator into apply the custom code
# regions to the actual final output version of this sub-sdk file.

# region imports
# endregion imports


# region sdk-class-body
    def custom_health_check(self) -> bool:
        res = self.check()
        if res.http_meta is None or res.http_meta.response is None:
            raise Exception("Expected res.http_meta.response to be set")
        return utils.match_status_codes(["2XX"], res.http_meta.response.status_code)


# endregion sdk-class-body

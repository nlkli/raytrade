# raytrade

[![rytrade demo 1 Sat Feb 28 01:31:02 MSK 2026]()](https://github.com/nlkli/assetsrepo/blob/main/raytrade.demo/demo1-o.mp4)  

[![rytrade demo 2 Sat Feb 28 01:31:16 MSK 2026]()](https://github.com/nlkli/assetsrepo/blob/main/raytrade.demo/demo2-o.mp4)

### Component params

rhd - row height diff

- split
    - axis (0 or 1)
    - s - size
    - mode - split mode (0, 1, 2)

- chart
    - rhd (app rh - chart rhd) = chart rh
    - sx - scale x axis
    - sy - scale y axis
    - tx - shift x
    - ty - shift y
    - show_lable (bool) - top left instrument info
    - show_grid (bool)

- orderbook
    - rhd (app rh - ob rhd) = ob rh
    - vm - view mode (0 or 1)
    - show_text (bool)

- orderbook_plus (split orderbook)
    - ```json
{
    "type": "orderbook_plus",
    "params": {},
    "a": {
        "type": "orderbook",
        "params": {
            "vm": 1,
            "rhd": 4
        }
    },
    "b": {
        "type": "orderbook",
        "params": {
            "vm": 0,
            "rhd": 4
        }
    }
}
    ```


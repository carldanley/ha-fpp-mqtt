# ha-fpp-mqtt

> Home Assistant Falcon Player MQTT

[![Build Status](https://ci.r1p.io/api/badges/carldanley/ha-fpp-mqtt/status.svg)](https://ci.r1p.io/carldanley/ha-fpp-mqtt)

## About

This is a project that monitors the state of the overlay models in any given controller running Falcon Player. The objective is to query the state of the overlay models on a regular interval and report status changes (if any). This project also accepts commands and translates them back into falcon player mqtt messages accordingly.

Another interesting thing is that Falcon Player doesn't currently have the ability to query for a specific overlay model's color but this project is also a state machine and will store the last known color (which helps fill in the gaps) for state-driven MQTT projects.

## Example JSON Payloads

### Turn on an Overlay Model

```json
{
  "state": "on"
}
```

### Turn off an Overlay Model

```json
{
  "state": "off"
}
```

### Turn on an Overlay Model and set the color to red

```json
{
  "state": "on",
  "color": [255, 255, 0]
}
```

### Status Update sent by this program

```json
{
  "state": "on",
  "color": [255, 0, 0]
}
```

## Home Assistant

Here is an example configuration for how to set up a light within Home Assistant.

```yaml
mqtt:
  light:
  - name: backyard_back_fence_palm_flood_middle
    schema: template
    state_topic: "alfred/ha-fpp-mqtt/backyard_back_fence_palm_flood_middle/status"
    state_template: "{{ value_json.state }}"
    command_topic: "alfred/ha-fpp-mqtt/backyard_back_fence_palm_flood_middle/set"
    command_off_template: >
      {
        "controller": "backyard-kulp-2",
        "state": "off"
      }
    command_on_template: >
      {
        "controller": "backyard-kulp-2",
        "state": "on"
        {%- if red is defined and green is defined and blue is defined -%}
        , "color": [{{ red }}, {{ green }}, {{ blue }}]
        {%- endif -%}
      }
    red_template: "{{ value_json.color[0] }}"
    green_template: "{{ value_json.color[1] }}"
    blue_template: "{{ value_json.color[2] }}"
```

## FAQs

### Why is brightness not supported?

This is simple: Falcon Player does not currently have an API call (or MQTT command) to set the brightness of a specific model. The brightness is handled when the controller configuration is either updated via Falcon Player or xLights.

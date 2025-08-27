#!/bin/bash

# Basic ANSI colors (0–7)
echo "https://en.wikipedia.org/wiki/ANSI_escape_code#3-bit_and_4-bit"
echo -e "\e[30m Black (0) \e[0m"
echo -e "\e[31m Red (1) \e[0m"
echo -e "\e[32m Green (2) \e[0m"
echo -e "\e[33m Yellow (3) \e[0m"
echo -e "\e[34m Blue (4) \e[0m"
echo -e "\e[35m Magenta (5) \e[0m"
echo -e "\e[36m Cyan (6) \e[0m"
echo -e "\e[37m White (7) \e[0m"

# Bright colors (8–15)
echo -e "\e[90m Bright Black (8) \e[0m"
echo -e "\e[91m Bright Red (9) \e[0m"
echo -e "\e[92m Bright Green (10) \e[0m"
echo -e "\e[93m Bright Yellow (11) \e[0m"
echo -e "\e[94m Bright Blue (12) \e[0m"
echo -e "\e[95m Bright Magenta (13) \e[0m"
echo -e "\e[96m Bright Cyan (14) \e[0m"
echo -e "\e[97m Bright White (15) \e[0m"

echo -e "\e[40m Black (0) \e[0m"
echo -e "\e[41m Red (1) \e[0m"
echo -e "\e[42m Green (2) \e[0m"
echo -e "\e[43m Yellow (3) \e[0m"
echo -e "\e[44m Blue (4) \e[0m"
echo -e "\e[45m Magenta (5) \e[0m"
echo -e "\e[46m Cyan (6) \e[0m"
echo -e "\e[47m White (7) \e[0m"

echo -e "\e[100m Bright Black (8) \e[0m"
echo -e "\e[101m Bright Red (9) \e[0m"
echo -e "\e[102m Bright Green (10) \e[0m"
echo -e "\e[103m Bright Yellow (11) \e[0m"
echo -e "\e[104m Bright Blue (12) \e[0m"
echo -e "\e[105m Bright Magenta (13) \e[0m"
echo -e "\e[106m Bright Cyan (14) \e[0m"
echo -e "\e[107m Bright White (15) \e[0m"

# 256-color foreground (some)
echo -e "\x1b[38;5;1m 256 Color Index 1 \x1b[0m"
echo -e "\x1b[38;5;46m 256 Color Index 46 (greenish) \x1b[0m"
echo -e "\x1b[38;5;196m 256 Color Index 196 (bright red) \x1b[0m"
echo -e "\x1b[38;5;220m 256 Color Index 220 (yellow-orange) \x1b[0m"
echo -e "\x1b[38;5;255m 256 Color Index 255 (white) \x1b[0m"

# With `:`
echo -e "\x1b[38:5:1m 256 Color Index 1 \x1b[0m"
echo -e "\x1b[38:5:46m 256 Color Index 46 (greenish) \x1b[0m"
echo -e "\x1b[38:5:196m 256 Color Index 196 (bright red) \x1b[0m"
echo -e "\x1b[38:5:220m 256 Color Index 220 (yellow-orange) \x1b[0m"
echo -e "\x1b[38:5:255m 256 Color Index 255 (white) \x1b[0m"

# 256-color  background (some)
echo -e "\x1b[48;5;1m 256 Color Index 1 \x1b[0m"
echo -e "\x1b[48;5;46m 256 Color Index 46 (greenish) \x1b[0m"
echo -e "\x1b[48;5;196m 256 Color Index 196 (bright red) \x1b[0m"
echo -e "\x1b[48;5;220m 256 Color Index 220 (yellow-orange) \x1b[0m"
echo -e "\x1b[48;5;255m 256 Color Index 255 (white) \x1b[0m"


# True color forground
echo -e "\x1b[38;2;255;0;0m RGB Red \x1b[0m"
echo -e "\x1b[38;2;0;255;0m RGB Green \x1b[0m"
echo -e "\x1b[38;2;0;0;255m RGB Blue \x1b[0m"
echo -e "\x1b[38;2;255;255;0m RGB Yellow \x1b[0m"
echo -e "\x1b[38;2;255;165;0m RGB Orange \x1b[0m"
echo -e "\x1b[38;2;128;0;128m RGB Purple \x1b[0m"

# True color background
echo -e "\x1b[48;2;255;0;0m Background RGB Red \x1b[0m"
echo -e "\x1b[48;2;0;255;0m Background RGB Green \x1b[0m"
echo -e "\x1b[48;2;0;0;255m Background RGB Blue \x1b[0m"

# Mixed foreground + background
echo -e "\x1b[38;2;255;255;255m\x1b[48;2;0;0;0m White on Black \x1b[0m"
echo -e "\x1b[38;2;0;0;0m\x1b[48;2;255;255;255m Black on White \x1b[0m"


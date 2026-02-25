import os

with open("internal/generator/generator.go", "r") as f:
    lines = f.readlines()

def extract(start_marker, end_marker, out_filename):
    start_idx = -1
    for i, line in enumerate(lines):
        if line.startswith(start_marker):
            start_idx = i
            break
    
    if start_idx == -1: return
    
    # The first line of the template has a backtick
    content_lines = []
    content_lines.append(lines[start_idx].split("`", 1)[1])
    
    # Now read until the closing backtick on a single line
    for i in range(start_idx + 1, len(lines)):
        if lines[i].strip() == "`":
            break
        content_lines.append(lines[i])
        
    with open("internal/generator/templates/" + out_filename, "w") as f:
        f.writelines(content_lines)

extract("const contractTemplate =", "contract.sol.tmpl")
extract("const deployTemplate =", "deploy.js.tmpl")
extract("const testTemplate =", "test.js.tmpl")

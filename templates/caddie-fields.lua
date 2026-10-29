-- Keeps the field lines under a `###` entry title (organisation, location,
-- dates; or venue, authors, date) on their own lines and hands them to
-- `caddie-fields` in typst-template.typ for styling. Formats only: no text
-- is added, removed or reordered.

local function hard_breaks(para)
  return para.content:walk({
    SoftBreak = function()
      return pandoc.LineBreak()
    end,
  })
end

function Blocks(blocks)
  local out = pandoc.Blocks({})
  for i, block in ipairs(blocks) do
    local prev = blocks[i - 1]
    local is_fields = block.t == "Para" and prev ~= nil
      and prev.t == "Header" and prev.level == 3
    if is_fields then
      out:insert(pandoc.RawBlock("typst", "#caddie-fields["))
      out:insert(pandoc.Para(hard_breaks(block)))
      out:insert(pandoc.RawBlock("typst", "]"))
    else
      out:insert(block)
    end
  end
  return out
end

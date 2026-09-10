data "testing_parameters_deep_object" "my_parametersdeepobject" {
  inline_object = {
    boolean_level_1 = true
    integer_level_1 = 7
    list_string_level_1 = [
      "..."
    ]
    object_level_1 = {
      boolean_level_2 = false
      integer_level_2 = 8
      list_string_level_2 = [
        "..."
      ]
      object_level_2 = {
        boolean_level_3 = false
        integer_level_3 = 4
        list_string_level_3 = [
          "..."
        ]
        string_level_3 = "...my_string_level_3..."
      }
      string_level_2 = "...my_string_level_2..."
    }
    string_level_1 = "...my_string_level_1..."
  }
  ref_object = {
    boolean_level_1 = false
    integer_level_1 = 6
    list_string_level_1 = [
      "..."
    ]
    object_level_1 = {
      boolean_level_2 = false
      integer_level_2 = 2
      list_string_level_2 = [
        "..."
      ]
      object_level_2 = {
        boolean_level_3 = false
        integer_level_3 = 6
        list_string_level_3 = [
          "..."
        ]
        string_level_3 = "...my_string_level_3..."
      }
      string_level_2 = "...my_string_level_2..."
    }
    string_level_1 = "...my_string_level_1..."
  }
}